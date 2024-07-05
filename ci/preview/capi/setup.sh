#!/bin/bash
set -euo pipefail

source common.sh

setup_kubectl

export OPENSTACK_CLOUD_YAML_B64=$(kubectl get secret cluster-admin-keystone -o 'jsonpath={.data.clouds\.yaml}')

log "Generating and applying Cluster API manifests to undercloud"
sed "s/\$(CLUSTER_NAME)/$CLUSTER_NAME/" cluster.yaml | \
    sed -e "s/\$(OPENSTACK_FAILURE_DOMAIN)/$OPENSTACK_FAILURE_DOMAIN/" | \
    kubectl apply -f-

log "Waiting for Kubernetes control plane to be available"
kubectl wait kubeadmcontrolplane/$CLUSTER_NAME-control-plane \
    --for=condition=Available \
    --timeout=10m

LB_SUBNET_ID=$(kubectl get openstackcluster $CLUSTER_NAME -o 'jsonpath={.status.network.subnet.id}')
EXTERNAL_NET_ID=$(kubectl get openstackcluster $CLUSTER_NAME -o 'jsonpath={.status.externalNetwork.id}')

echo $OPENSTACK_CLOUD_YAML_B64 | base64 -d > clouds.yaml

cat <<EOF > cloud.conf
[Global]
auth-url=$(yq eval .clouds.default.auth.auth_url clouds.yaml)
region=$(yq eval .clouds.default.region_name clouds.yaml)
username=$(yq eval .clouds.default.auth.username clouds.yaml)
password=$(yq eval .clouds.default.auth.password clouds.yaml)
tenant-name=$(yq eval .clouds.default.auth.project_name clouds.yaml)
domain-name=$(yq eval .clouds.default.auth.user_domain_name clouds.yaml)

[LoadBalancer]
subnet-id=$LB_SUBNET_ID
floating-network-id=$EXTERNAL_NET_ID

[BlockStorage]
bs-version=v2
ignore-volume-az=true

[LoadBalancer]
lb-provider=ovn
lb-method=SOURCE_IP_PORT
EOF

log "Switching kubectl to $CLUSTER_NAME cluster"
clusterctl get kubeconfig $CLUSTER_NAME > kubeconfig
export KUBECONFIG=$(pwd)/kubeconfig

kubectl config set-context --current --namespace default

log "Ensuring kube-system/cloud-config secret"
kubectl -n kube-system get secret cloud-config && kubectl -n kube-system delete secret cloud-config
kubectl -n kube-system create secret generic cloud-config --from-file=cloud.conf

rm cloud.conf clouds.yaml

log "Applying cloud provider manifests"
kubectl kustomize cloud-provider | \
    sed "s/\$(CLUSTER_NAME)/$CLUSTER_NAME/" | \
    kubectl apply -f-

# Install cluster networking
log "Applying Calico operator manifests"
kubectl apply -f https://raw.githubusercontent.com/projectcalico/calico/v3.26.4/manifests/tigera-operator.yaml --server-side=true

kubectl -n tigera-operator rollout status deployment tigera-operator

log "Applying Calico networking manifests"
kubectl apply -f calico.yaml

log "Waiting for Calico to be ready"
# NOTE it can take awhile until the calico operator creates the deployment
attempts=0
until kubectl -n calico-system get deploy calico-kube-controllers; do
    attempts=$((attempts + 1))
    if [ $attempts -lt 10 ]; then
        sleep 3
    else
        exit 1
    fi
done

kubectl -n calico-system rollout status deploy calico-kube-controllers

log "Applying cinder manifests"
kubectl kustomize cinder-csi | \
    sed "s/\$(CLUSTER_NAME)/$CLUSTER_NAME/" | \
    kubectl apply -f-

log "Applying ingress-nginx manifests"
curl -sL https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.10.0/deploy/static/provider/cloud/deploy.yaml | \
    # Octavia does not work with Local policy
    sed -e "s/externalTrafficPolicy: Local/externalTrafficPolicy: Cluster/g" | \
    kubectl apply -f -

log "Applying cert-manager manifests"
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.14.3/cert-manager.yaml

log "Ensuring cert-manager/google-dns secret"
if ! kubectl -n cert-manager get secret google-dns; then
    echo $GCLOUD_SVC_ACCOUNT_KEY_JSON | base64 -d > key
    kubectl -n cert-manager create secret generic google-dns --from-file=key
    rm key
fi

log "Waiting for cert-manager to become ready"
kubectl -n cert-manager rollout status deploy

log "Applying cert-manager issuer and wildcard cert"
# NOTE cert-manager webhook can take a few seconds to become ready
attempts=0
until kubectl apply -f issuer-letsencrypt.yaml; do
    attempts=$((attempts + 1))
    if [ $attempts -lt 10 ]; then
        sleep 3
    else
        exit 1
    fi
done

sed "s/\$(CLUSTER_DOMAIN)/$CLUSTER_DOMAIN/" certificate.yaml | kubectl apply -f-

log "Applying openstack-operator manifests"
kustomize build ../../../config/default | kubectl apply -f-

log "Waiting for ingress load balancer to become ready"
ingress_lb_ip=""
while [ -z $ingress_lb_ip ]; do
    ingress_lb_ip=$(kubectl -n ingress-nginx get svc ingress-nginx-controller -o 'jsonpath={.status.loadBalancer.ingress[0].ip}')
    [ -z "$ingress_lb_ip" ] && sleep 5
done

log "Upserting wildcard DNS record for ingress"
if gcloud dns record-sets describe "*.$CLUSTER_DOMAIN" --type=A --zone=$DNS_ZONE; then
    gcloud dns record-sets update "*.$CLUSTER_DOMAIN" --type=A --zone=$DNS_ZONE --rrdatas=$ingress_lb_ip --ttl=300
else
    gcloud dns record-sets create "*.$CLUSTER_DOMAIN" --type=A --zone=$DNS_ZONE --rrdatas=$ingress_lb_ip --ttl=300
fi

log "Waiting for openstack-operator to become ready"
kubectl -n openstack-system rollout status deploy openstack-operator-controller-manager

log "Applying OpenStack control plane manifests"
sed "s/\$(CLUSTER_DOMAIN)/$CLUSTER_DOMAIN/" controlplane.yaml | kubectl apply -f-

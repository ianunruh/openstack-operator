#!/bin/bash
set -euo pipefail

source common.sh

setup_kubectl

export OPENSTACK_CLOUD="default"
export OPENSTACK_CLOUD_CACERT_B64=$(kubectl get secret cluster-admin-keystone -o 'jsonpath={.data.cacert}')
export OPENSTACK_CLOUD_YAML_B64=$(kubectl get secret cluster-admin-keystone -o 'jsonpath={.data.clouds\.yaml}')
export OPENSTACK_DNS_NAMESERVERS="1.1.1.1"
export OPENSTACK_IMAGE_NAME="ubuntu-2004-kube-v1.20.9"
export OPENSTACK_SSH_KEY_NAME="ianunruh-yubikey"

export OPENSTACK_CONTROL_PLANE_MACHINE_FLAVOR="c1-medium"
export OPENSTACK_NODE_MACHINE_FLAVOR="c1-large"

log "Generating and applying Cluster API manifests to undercloud"
clusterctl generate cluster $CLUSTER_NAME \
    --kubernetes-version 1.20.9 \
    --control-plane-machine-count 1 \
    --worker-machine-count 3 \
    --flavor external-cloud-provider \
    | kubectl apply -f-

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
EOF

log "Switching kubectl to $CLUSTER_NAME cluster"
clusterctl get kubeconfig $CLUSTER_NAME > kubeconfig
export KUBECONFIG=$(pwd)/kubeconfig

kubectl config set-context --current --namespace default

log "Ensuring kube-system/cloud-config secret"
kubectl -n kube-system get secret cloud-config && kubectl -n kube-system delete secret cloud-config
kubectl -n kube-system create secret generic cloud-config --from-file=cloud.conf

log "Excluding control plane nodes from load balancer"
kubectl label node node.kubernetes.io/exclude-from-external-load-balancers= \
    -l node-role.kubernetes.io/control-plane= \
    --overwrite

# Install cluster networking
log "Applying Calico networking manifests"
kubectl apply -f https://docs.projectcalico.org/v3.18/manifests/calico.yaml

log "Applying cloud provider manifests"
kubectl apply -f https://raw.githubusercontent.com/kubernetes/cloud-provider-openstack/v1.22.0/manifests/controller-manager/cloud-controller-manager-roles.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes/cloud-provider-openstack/v1.22.0/manifests/controller-manager/cloud-controller-manager-role-bindings.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes/cloud-provider-openstack/v1.22.0/manifests/controller-manager/openstack-cloud-controller-manager-ds.yaml

log "Applying Cinder CSI driver manifests"
kubectl apply -f https://raw.githubusercontent.com/kubernetes/cloud-provider-openstack/v1.22.0/manifests/cinder-csi-plugin/cinder-csi-controllerplugin-rbac.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes/cloud-provider-openstack/v1.22.0/manifests/cinder-csi-plugin/cinder-csi-controllerplugin.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes/cloud-provider-openstack/v1.22.0/manifests/cinder-csi-plugin/cinder-csi-nodeplugin-rbac.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes/cloud-provider-openstack/v1.22.0/manifests/cinder-csi-plugin/cinder-csi-nodeplugin.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes/cloud-provider-openstack/v1.22.0/manifests/cinder-csi-plugin/csi-cinder-driver.yaml

kubectl apply -f cinder-storageclass.yaml

log "Applying ingress-nginx manifests"
curl -sL https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.0.0/deploy/static/provider/cloud/deploy.yaml -o ingress.yaml
# Octavia does not work with Local policy
sed -i.bak "s/externalTrafficPolicy: Local/externalTrafficPolicy: Cluster/g" ingress.yaml
kubectl apply -f ingress.yaml

log "Applying cert-manager manifests"
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.5.3/cert-manager.yaml

log "Ensuring cert-manager/google-dns secret"
if ! kubectl -n cert-manager get secret google-dns; then
    echo $GCLOUD_SVC_ACCOUNT_KEY_JSON | base64 -d > key
    kubectl -n cert-manager create secret generic google-dns --from-file=key
    rm key
fi

log "Waiting for cert-manager to become ready"
kubectl -n cert-manager rollout status deploy cert-manager

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
kustomize build ../../config/default | kubectl apply -f-

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

#!/bin/bash
set -ex

CLUSTER_DOMAIN=$CLUSTER_NAME.$CLOUDSDK_COMPUTE_ZONE.test.ospk8s.com

if ! gcloud container clusters describe $CLUSTER_NAME >/dev/null; then
    gcloud container clusters create $CLUSTER_NAME \
        --num-nodes "3" \
        --cluster-version "1.20.9-gke.701" \
        --release-channel "regular" \
        --machine-type "e2-standard-2" \
        --image-type "UBUNTU_CONTAINERD" \
        --metadata disable-legacy-endpoints=true \
        --scopes "https://www.googleapis.com/auth/devstorage.read_only","https://www.googleapis.com/auth/logging.write","https://www.googleapis.com/auth/monitoring","https://www.googleapis.com/auth/servicecontrol","https://www.googleapis.com/auth/service.management.readonly","https://www.googleapis.com/auth/trace.append" \
        --create-subnetwork "" \
        --enable-ip-alias \
        --no-enable-intra-node-visibility \
        --no-enable-master-authorized-networks \
        --no-enable-basic-auth \
        --default-max-pods-per-node "110" \
        --max-pods-per-node "110" \
        --addons HorizontalPodAutoscaling,HttpLoadBalancing,GcePersistentDiskCsiDriver \
        --enable-shielded-nodes
else
    gcloud container clusters get-credentials $CLUSTER_NAME
fi

# Install ingress-nginx
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.0.0/deploy/static/provider/cloud/deploy.yaml

# Install cert-manager
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.5.3/cert-manager.yaml

if ! kubectl -n cert-manager get secret google-dns; then
    echo $GCLOUD_SVC_ACCOUNT_KEY_JSON | base64 -d > key
    kubectl -n cert-manager create secret generic google-dns --from-file=key
    rm key
fi

kubectl -n cert-manager rollout status deploy cert-manager

# Configure letsencrypt issuer and wildcard TLS cert
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

# Install operator
kubectl apply -k ../../config/default

# Get the ingress-nginx load balancer IP
ingress_lb_ip=""
while [ -z $ingress_lb_ip ]; do
    echo "Waiting for ingress load balancer..."
    ingress_lb_ip=$(kubectl -n ingress-nginx get svc ingress-nginx-controller -o 'jsonpath={.status.loadBalancer.ingress[0].ip}')
    [ -z "$ingress_lb_ip" ] && sleep 5
done

# Upsert wildcard DNS pointing at ingress
if gcloud dns record-sets describe "*.$CLUSTER_DOMAIN" --type=A --zone=$DNS_ZONE; then
    gcloud dns record-sets update "*.$CLUSTER_DOMAIN" --type=A --zone=$DNS_ZONE --rrdatas=$ingress_lb_ip --ttl=300
else
    gcloud dns record-sets create "*.$CLUSTER_DOMAIN" --type=A --zone=$DNS_ZONE --rrdatas=$ingress_lb_ip --ttl=300
fi

# Wait for operator to be available
kubectl -n openstack-system rollout status deploy openstack-operator-controller-manager

# Configure OpenStack control plane
sed "s/\$(CLUSTER_DOMAIN)/$CLUSTER_DOMAIN/" controlplane.yaml | kubectl apply -f-

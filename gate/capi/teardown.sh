#!/bin/bash
set -euo pipefail

source common.sh

setup_kubectl

log "Switching kubectl to $CLUSTER_NAME cluster"
clusterctl get kubeconfig $CLUSTER_NAME > kubeconfig
export KUBECONFIG=$(pwd)/kubeconfig

kubectl config set-context --current --namespace default

log "Tearing down OpenStack control plane"
kubectl delete controlplane default

log "Cleaning up volumes"
kubectl delete pvc --all

log "Cleaning up ingress load balancer"
kubectl -n ingress-nginx delete svc ingress-nginx-controller

# Switch back to undercluster
unset KUBECONFIG

log "Tearing down $CLUSTER_NAME Kubernetes cluster"
kubectl delete cluster $CLUSTER_NAME

log "Cleaning up cluster secrets"
kubectl delete secret $CLUSTER_NAME-cloud-config

log "Cleaning up ingress wildcard DNS record"
if gcloud dns record-sets describe "*.$CLUSTER_DOMAIN" --type=A --zone=$DNS_ZONE; then
    gcloud dns record-sets delete "*.$CLUSTER_DOMAIN" --type=A --zone=$DNS_ZONE
fi

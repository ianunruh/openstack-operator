#!/bin/bash
set -euo pipefail

source common.sh

setup_kubectl
setup_openstack

log "Switching kubectl to $CLUSTER_NAME cluster"
clusterctl get kubeconfig $CLUSTER_NAME > kubeconfig
export KUBECONFIG=$(pwd)/kubeconfig

kubectl config set-context --current --namespace default

log "Cleaning up ingress load balancer"
kubectl -n ingress-nginx delete svc --all

# Switch back to undercluster
unset KUBECONFIG

log "Tearing down $CLUSTER_NAME Kubernetes cluster"
kubectl delete cluster $CLUSTER_NAME

log "Cleaning up Cinder volumes"
for volume_id in $(openstack volume list --long -f json | yq -r ".[]|select(.Properties[\"cinder.csi.openstack.org/cluster\"] == \"$CLUSTER_NAME\").ID"); do
    for attachment_id in $(openstack volume attachment list --volume-id=$volume_id -f json | yq -r '.[].ID'); do
        log "Deleting volume attachment $attachment_id"
        openstack volume attachment delete $attachment_id
    done
    log "Deleting volume $volume_id"
    openstack volume delete $volume_id
done

log "Cleaning up ingress wildcard DNS record"
if gcloud dns record-sets describe "*.$CLUSTER_DOMAIN" --type=A --zone=$DNS_ZONE; then
    gcloud dns record-sets delete "*.$CLUSTER_DOMAIN" --type=A --zone=$DNS_ZONE
fi

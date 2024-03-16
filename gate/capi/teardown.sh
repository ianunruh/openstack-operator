#!/bin/bash
set -euo pipefail

source common.sh

setup_kubectl
setup_openstack

log "Switching kubectl to $CLUSTER_NAME cluster"
clusterctl get kubeconfig $CLUSTER_NAME > kubeconfig
export KUBECONFIG=$(pwd)/kubeconfig

kubectl config set-context --current --namespace default

log "Tearing down OpenStack control plane"
kubectl get crd controlplanes.openstack.ospk8s.com && kubectl delete controlplane --all

kubectl wait pod --for delete --all

log "Cleaning up volumes"
kubectl delete pvc --all

log "Cleaning up ingress load balancer"
kubectl -n ingress-nginx delete svc --all

# Switch back to undercluster
unset KUBECONFIG

log "Waiting for all volumes to be cleaned up"
while true; do
    output=$(openstack volume list --long -f json | \
        yq ".[]|select(.Properties[\"cinder.csi.openstack.org/cluster\"]==\"$CLUSTER_NAME\")")
    if [[ "$output" -eq 0 ]]; then
        break
    fi
    log "Waiting 5 more seconds for all volumes to be cleaned up"
    sleep 5
done

log "Waiting for all load balancers to be cleaned up"
while true; do
    output=$(openstack loadbalancer list -f json | \
        yq ".[]|select(.name|contains(\"kube_service_${CLUSTER_NAME}_\"))")
    if [[ "$output" -eq 0 ]]; then
        break
    fi
    log "Waiting 5 more seconds for all load balancers to be cleaned up"
    sleep 5
done

log "Tearing down $CLUSTER_NAME Kubernetes cluster"
kubectl delete cluster $CLUSTER_NAME

log "Cleaning up ingress wildcard DNS record"
if gcloud dns record-sets describe "*.$CLUSTER_DOMAIN" --type=A --zone=$DNS_ZONE; then
    gcloud dns record-sets delete "*.$CLUSTER_DOMAIN" --type=A --zone=$DNS_ZONE
fi

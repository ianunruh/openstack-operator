#!/bin/bash
set -euxo pipefail

CLUSTER_DOMAIN=$CLUSTER_NAME.$OPENSTACK_FAILURE_DOMAIN.test.ospk8s.com

# Clean up LBs/floating IPs
kubectl -n ingress-nginx delete svc ingress-nginx-controller

kubectl delete cluster $CLUSTER_NAME

# TODO clean up cinder volumes

if gcloud dns record-sets describe "*.$CLUSTER_DOMAIN" --type=A --zone=$DNS_ZONE; then
    gcloud dns record-sets delete "*.$CLUSTER_DOMAIN" --type=A --zone=$DNS_ZONE
fi

# Clean up secrets
kubectl delete secret $CLUSTER_NAME-cloud-config

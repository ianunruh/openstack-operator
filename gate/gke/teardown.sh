#!/bin/bash
set -ex

CLUSTER_DOMAIN=$CLUSTER_NAME.$CLOUDSDK_COMPUTE_ZONE.test.ospk8s.com

if gcloud container clusters describe $CLUSTER_NAME >/dev/null; then
    gcloud container clusters delete $CLUSTER_NAME
fi

if gcloud dns record-sets describe "*.$CLUSTER_DOMAIN" --type=A --zone=$DNS_ZONE; then
    gcloud dns record-sets delete "*.$CLUSTER_DOMAIN" --type=A --zone=$DNS_ZONE
else

for disk in $(gcloud compute disks list | grep gke-$CLUSTER_NAME | awk '{print $1}'); do
    gcloud compute disks delete $disk
done

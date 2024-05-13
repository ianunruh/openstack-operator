#!/bin/bash
set -eu -o pipefail

CERT_PATH=/tmp/k8s-webhook-server/serving-certs

mkdir -p "$CERT_PATH"

kubectl -n openstack-system scale deploy openstack-operator-controller-manager --replicas 0

kubectl -n openstack-system get secret webhook-server-cert \
    -o 'jsonpath={.data.ca\.crt}' \
    | base64 -d > "$CERT_PATH/ca.crt"

kubectl -n openstack-system get secret webhook-server-cert \
    -o 'jsonpath={.data.tls\.crt}' \
    | base64 -d > "$CERT_PATH/tls.crt"

kubectl -n openstack-system get secret webhook-server-cert \
    -o 'jsonpath={.data.tls\.key}' \
    | base64 -d > "$CERT_PATH/tls.key"

kubectl get mutatingwebhookconfiguration openstack-operator-mutating-webhook-configuration -o json \
    | hack/webhook-patch.py \
    | kubectl apply -f -

kubectl get validatingwebhookconfiguration openstack-operator-validating-webhook-configuration -o json \
    | hack/webhook-patch.py \
    | kubectl apply -f -

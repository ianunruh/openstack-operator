#!/bin/bash
set -euo pipefail

source common.sh

setup_kubectl

log "Switching kubectl to $CLUSTER_NAME cluster"
clusterctl get kubeconfig $CLUSTER_NAME > kubeconfig
export KUBECONFIG=$(pwd)/kubeconfig

kubectl config set-context --current --namespace default

log "Ensuring OAuth app set up"
if ! kubectl get secret keystone-oidc; then
    oidc_redirect_uri=https://keystone.$CLUSTER_DOMAIN/v3/OS-FEDERATION/identity_providers/gitlab/protocols/openid/auth

    gitlab_app=$(curl -s -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
        -d "name=$CLUSTER_NAME&redirect_uri=$oidc_redirect_uri&scopes=read_user openid profile email" \
        "https://gitlab.kcloud.io/api/v4/applications")

    kubectl create secret generic keystone-oidc \
        --from-literal=KEYSTONE_OIDC_CLIENT_ID=$(echo $gitlab_app | yq -r '.application_id') \
        --from-literal=KEYSTONE_OIDC_CLIENT_SECRET=$(echo $gitlab_app | yq -r '.secret') \
        --from-literal=KEYSTONE_OIDC_CRYPTO_PASSPHRASE=$(python -c 'import secrets; print(secrets.token_hex(24))')
fi

log "Waiting for keystone-api to become ready"
kubectl rollout status deploy keystone-api

setup_openstack keystone

log "Ensuring Keystone federation set up"
if ! openstack group show federated_users --domain default; then
    openstack group create federated_users --domain default
fi

openstack role add admin --group federated_users --project admin

if ! openstack identity provider show gitlab; then
    openstack identity provider create gitlab --remote-id https://gitlab.kcloud.io
fi

if ! openstack mapping show gitlab; then
    openstack mapping create gitlab --rules oidc-mapping-rules.json
fi

if ! openstack federation protocol show openid --identity-provider gitlab; then
    openstack federation protocol create openid --identity-provider gitlab --mapping gitlab
fi

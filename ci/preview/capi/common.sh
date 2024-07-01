#!/bin/bash
set -eu

CLUSTER_DOMAIN=$CLUSTER_NAME.$OPENSTACK_FAILURE_DOMAIN.test.ospk8s.com

export TERM=xterm-color

blue=$(tput setaf 6)
reset=$(tput sgr0)

log() {
    echo "${blue}$(date -u) [INFO] $1${reset}" >&2
}

setup_kubectl() {
    mkdir -p $HOME/.kube
    if [ ! -f $HOME/.kube/config ]; then
        echo $PREVIEW_KUBECONFIG | base64 -d > $HOME/.kube/config
    fi

    log "Switching kubectl to CI namespace"
    kubectl config set-context --current --namespace ospk8s-ci
}

export OS_CLOUD="default"

setup_openstack() {
    mkdir -p $HOME/.config/openstack
    kubectl get secret $1 -o 'jsonpath={.data.clouds\.yaml}' | base64 -d > $HOME/.config/openstack/clouds.yaml
    export OS_VOLUME_API_VERSION=3.33
}

openstack() {
    pipenv run openstack $@
}

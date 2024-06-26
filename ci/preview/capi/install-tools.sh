#!/bin/bash
set -eux

mkdir -p $HOME/.local/bin

if ! [ -f $HOME/.local/bin/kubectl ]; then
    curl -sL https://dl.k8s.io/release/v1.26.7/bin/linux/amd64/kubectl -o kubectl
    chmod +x kubectl
    mv kubectl $HOME/.local/bin/kubectl
fi

if ! [ -f $HOME/.local/bin/kustomize ]; then
    curl -sL https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv5.3.0/kustomize_v5.3.0_linux_amd64.tar.gz -o kustomize.tar.gz
    tar xfv kustomize.tar.gz
    mv kustomize $HOME/.local/bin/kustomize
fi

if ! [ -f $HOME/.local/bin/clusterctl ]; then
    curl -sL https://github.com/kubernetes-sigs/cluster-api/releases/download/v1.6.2/clusterctl-linux-amd64 -o clusterctl
    chmod +x clusterctl
    mv clusterctl $HOME/.local/bin/clusterctl
fi

if ! [ -f $HOME/.local/bin/yq ]; then
    curl -sL https://github.com/mikefarah/yq/releases/download/v4.40.5/yq_linux_amd64 -o yq
    chmod +x yq
    mv yq $HOME/.local/bin/yq
fi

pip install pipenv

pipenv install

#!/bin/bash
set -eu

(cd controlplane && kustomize edit add patch --path ../profiles/tls-disabled/controlplane.yaml)

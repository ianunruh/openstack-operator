#!/bin/bash
set -eu

(cd cluster && kustomize edit add patch --path ../profiles/ha/cluster.yaml)
(cd controlplane && kustomize edit add patch --path ../profiles/ha/controlplane.yaml)

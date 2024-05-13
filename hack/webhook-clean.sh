#!/bin/bash
set -eu

kubectl delete mutatingwebhookconfiguration openstack-operator-mutating-webhook-configuration
kubectl delete validatingwebhookconfiguration openstack-operator-validating-webhook-configuration

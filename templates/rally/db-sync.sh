#!/bin/bash
set -ex

revision=$(rally db revision)
if [ $revision = "None" ]; then
    rally db create
else
    rally db upgrade
fi

# Use public Keystone endpoint
export OS_AUTH_URL=$OS_AUTH_URL_WWW

if ! rally deployment list | grep openstack; then
    rally deployment create --fromenv --name openstack
fi

#!/bin/bash
set -ex

: ${RALLY_DEPLOYMENT_NAME:=openstack}
: ${RALLY_TASK_PATH:=samples/tasks/scenarios/keystone/create-user-update-password.yaml}

if ! rally deployment list | grep ${RALLY_DEPLOYMENT_NAME}; then
    rally deployment create --fromenv --name ${RALLY_DEPLOYMENT_NAME}
fi

rally deployment use ${RALLY_DEPLOYMENT_NAME}
rally deployment check

rally task start ${RALLY_TASK_PATH}

rally task sla-check

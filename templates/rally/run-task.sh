#!/bin/bash
set -ex

rally deployment use openstack
rally deployment check

rally task start ${RALLY_TASK_PATH}

rally task sla-check

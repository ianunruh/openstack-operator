#!/bin/bash
set -ex

neutron-db-manage \
    --config-file /etc/neutron/neutron.conf \
    upgrade head

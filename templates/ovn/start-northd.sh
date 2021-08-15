#!/bin/bash
set -ex

mkdir -p /var/run/ovn

exec ovn-northd \
    --ovnnb-db=$OVN_NB_CONNECTION \
    --ovnsb-db=$OVN_SB_CONNECTION

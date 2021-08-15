#!/bin/bash
set -ex

exec /usr/share/ovn/scripts/ovn-ctl run_nb_ovsdb \
    --db-nb-create-insecure-remote=yes \
    --ovn-nb-log="-vconsole:info -vfile:off"

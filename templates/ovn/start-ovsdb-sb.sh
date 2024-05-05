#!/bin/bash
set -eu

exec /usr/share/ovn/scripts/ovn-ctl run_sb_ovsdb --db-sb-create-insecure-remote=yes --ovn-sb-log="-vconsole:info -vfile:off"

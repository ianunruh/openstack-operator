#!/bin/bash
set -eu

exec ovn-northd --ovnnb-db=$OVN_NB_CONNECTION --ovnsb-db=$OVN_SB_CONNECTION

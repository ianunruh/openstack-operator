#!/bin/bash
set -ex

mkdir -p /var/run/ovn

exec ovn-controller

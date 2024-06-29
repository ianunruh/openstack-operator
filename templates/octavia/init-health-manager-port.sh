#!/bin/bash
set -eux -o pipefail

export OS_CLOUD=default

PORT_INFO=$(openstack port list --name octavia-health-manager-$HOSTNAME --network $HM_NETWORK_ID -f json)

HM_PORT_ID=$(echo $PORT_INFO | python3 -c 'import json, sys; print(json.load(sys.stdin)[0]["ID"])')
HM_PORT_MAC=$(echo $PORT_INFO | python3 -c 'import json, sys; print(json.load(sys.stdin)[0]["MAC Address"])')
HM_BIND_IP=$(echo $PORT_INFO | python3 -c 'import json, sys; print(json.load(sys.stdin)[0]["Fixed IP Addresses"][0]["ip_address"])')

echo $HM_PORT_ID > /tmp/pod-shared/HM_PORT_ID
echo $HM_PORT_MAC > /tmp/pod-shared/HM_PORT_MAC

cat > /tmp/pod-shared/octavia-health-manager.conf <<EOF
[health_manager]
bind_ip = $HM_BIND_IP
EOF

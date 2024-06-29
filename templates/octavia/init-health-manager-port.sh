#!/bin/bash
set -eux -o pipefail

export OS_CLOUD=default

PORT_INFO=$(openstack port list --name octavia-health-manager-$HOSTNAME --network $HM_NETWORK_ID -f json)

HM_PORT_ID=$(echo $PORT_INFO | python3 -c 'import json, sys; print(json.load(sys.stdin)[0]["ID"])')
HM_PORT_MAC=$(echo $PORT_INFO | python3 -c 'import json, sys; print(json.load(sys.stdin)[0]["MAC Address"])')
HM_BIND_IP=$(echo $PORT_INFO | python3 -c 'import json, sys; print(json.load(sys.stdin)[0]["Fixed IP Addresses"][0]["ip_address"])')

HM_IFACE=o-hm0

ovs-vsctl show

ovs-vsctl --may-exist add-port br-int ${HM_IFACE} \
    -- set Interface ${HM_IFACE} type=internal \
    -- set Interface ${HM_IFACE} external-ids:iface-status=active \
    -- set Interface ${HM_IFACE} external-ids:attached-mac=${HM_PORT_MAC} \
    -- set Interface ${HM_IFACE} external-ids:iface-id=${HM_PORT_ID} \
    -- set Interface ${HM_IFACE} external-ids:skip_cleanup=true

ip link set dev ${HM_IFACE} address ${HM_PORT_MAC}

cat > /tmp/pod-shared/octavia-health-manager.conf <<EOF
[health_manager]
bind_ip = $HM_BIND_IP
EOF

cat > /etc/dhcp/dhclient.conf <<EOF
request subnet-mask, broadcast-address, interface-mtu;
do-forward-updates false;
EOF

cat > /etc/dhcp/dhclient-enter-hooks.d/ignore-options <<EOF
unset new_dhcp_lease_time
unset new_domain_name new_domain_name_servers new_domain_search
unset new_rfc3442_classless_static_routes new_routers new_static_routes
EOF

dhclient -1 -v ${HM_IFACE}

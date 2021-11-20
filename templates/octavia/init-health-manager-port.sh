#!/bin/bash
set -eux

HM_IFACE=o-hm0

ovs-vsctl show

ovs-vsctl --may-exist add-port br-int ${HM_IFACE} \
    -- set Interface ${HM_IFACE} type=internal \
    -- set Interface ${HM_IFACE} external-ids:iface-status=active \
    -- set Interface ${HM_IFACE} external-ids:attached-mac=${HM_PORT_MAC} \
    -- set Interface ${HM_IFACE} external-ids:iface-id=${HM_PORT_ID} \
    -- set Interface ${HM_IFACE} external-ids:skip_cleanup=true

ip link set dev ${HM_IFACE} address ${HM_PORT_MAC}

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

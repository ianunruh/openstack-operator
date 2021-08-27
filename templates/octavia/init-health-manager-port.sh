#!/bin/bash
set -ex

HM_IFACE=o-hm0

ovs-vsctl show

ovs-vsctl --may-exist add-port br-int ${HM_IFACE} \
    -- set Interface ${HM_IFACE} type=internal \
    -- set Interface ${HM_IFACE} external-ids:iface-status=active \
    -- set Interface ${HM_IFACE} external-ids:attached-mac=${HM_PORT_MAC} \
    -- set Interface ${HM_IFACE} external-ids:iface-id=${HM_PORT_ID} \
    -- set Interface ${HM_IFACE} external-ids:skip_cleanup=true

ip link set dev ${HM_IFACE} address ${HM_PORT_MAC}

cat > /tmp/dhclient.conf <<EOF
request subnet-mask,broadcast-address,interface-mtu;
do-forward-updates false;
EOF

dhclient -v ${HM_IFACE} -cf /tmp/dhclient.conf

# prevent addr from expiring
HM_PREFIX=$(ip -4 addr show ${HM_IFACE} | awk '/inet/{print $2}')
ip addr change ${HM_PREFIX} dev ${HM_IFACE} valid_lft forever

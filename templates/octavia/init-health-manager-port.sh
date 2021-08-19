#!/bin/bash
set -ex

ovs-vsctl show

ovs-vsctl --may-exist add-port br-int o-hm0 \
    -- set Interface o-hm0 type=internal \
    -- set Interface o-hm0 external-ids:iface-status=active \
    -- set Interface o-hm0 external-ids:attached-mac=${HM_PORT_MAC} \
    -- set Interface o-hm0 external-ids:iface-id=${HM_PORT_ID} \
    -- set Interface o-hm0 external-ids:skip_cleanup=true

ip link set dev o-hm0 address ${HM_PORT_MAC}

cat > /tmp/dhclient.conf <<EOF
request subnet-mask,broadcast-address,interface-mtu;
do-forward-updates false;
EOF

dhclient -v o-hm0 -cf /tmp/dhclient.conf

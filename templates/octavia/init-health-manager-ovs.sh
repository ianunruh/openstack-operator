#!/bin/bash
set -eux -o pipefail

HM_PORT_ID=$(cat /tmp/pod-shared/HM_PORT_ID)
HM_PORT_MAC=$(cat /tmp/pod-shared/HM_PORT_MAC)

ovs-vsctl --may-exist add-port br-int ${HM_IFACE} \
    -- set Interface ${HM_IFACE} type=internal \
    -- set Interface ${HM_IFACE} external-ids:iface-status=active \
    -- set Interface ${HM_IFACE} external-ids:attached-mac=${HM_PORT_MAC} \
    -- set Interface ${HM_IFACE} external-ids:iface-id=${HM_PORT_ID} \
    -- set Interface ${HM_IFACE} external-ids:skip_cleanup=true

ip link set dev ${HM_IFACE} address ${HM_PORT_MAC}

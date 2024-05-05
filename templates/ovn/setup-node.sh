#!/bin/bash
set -eu

OVN_NODE_IP=$(python3 /scripts/get-encap-ip.py)

ovs-vsctl set open . external_ids:hostname=${HOSTNAME}
ovs-vsctl set open . external_ids:ovn-bridge=br-int
ovs-vsctl set open . external_ids:ovn-remote=${OVN_SB_CONNECTION}
ovs-vsctl set open . external_ids:ovn-encap-type=geneve
ovs-vsctl set open . external_ids:ovn-encap-ip=${OVN_NODE_IP}

if [ "${GATEWAY}" == "true" ]; then
    # mark it as gateway
    ovs-vsctl set open . external_ids:ovn-cms-options=enable-chassis-as-gw
    ovs-vsctl set open . external_ids:ovn-bridge-mappings=${BRIDGE_MAPPINGS}

    # setup bridges
    for bridge_port in ${BRIDGE_PORTS//,/ }
    do
        gw_bridge=$(echo ${bridge_port} | cut -d":" -f1)
        gw_port=$(echo ${bridge_port} | cut -d":" -f2)
        ovs-vsctl --may-exist add-br ${gw_bridge}
        ovs-vsctl --may-exist add-port ${gw_bridge} ${gw_port}
    done
fi

#!/bin/bash
set -ex

function quit {
    # Don't allow ovs-vswitchd to clear datapath flows on exit
    kill -9 $(cat /var/run/openvswitch/ovs-vswitchd.pid 2>/dev/null) 2>/dev/null || true
    kill $(cat /var/run/openvswitch/ovsdb-server.pid 2>/dev/null) 2>/dev/null || true
    exit 0
}
trap quit SIGTERM

/usr/share/openvswitch/scripts/ovs-ctl start --system-id=random

# ovs-appctl vlog/set "file:${OVS_LOG_LEVEL}"
/usr/share/openvswitch/scripts/ovs-ctl --protocol=udp --dport=6081 enable-protocol

# TODO wait for vswitchd to start
sleep 5

NIC=`ip -4 route | grep default | awk 'BEGIN{FS="dev "}{print $2}' | cut -d" " -f1`
OVN_NODE_IP=`ip -4 -o addr show "${NIC}" | awk 'BEGIN{FS="inet "}{print $2}' | cut -d" " -f1 | cut -d"/" -f1`

ovs-vsctl set open . external_ids:hostname=${HOSTNAME}
ovs-vsctl set open . external_ids:ovn-bridge=br-int
ovs-vsctl set open . external_ids:ovn-remote=${OVN_SB_CONNECTION}
ovs-vsctl set open . external_ids:ovn-encap-type=geneve
ovs-vsctl set open . external_ids:ovn-encap-ip=${OVN_NODE_IP}

if ${GATEWAY}; then
    # mark it as gateway
    ovs-vsctl set open . external_ids:ovn-cms-options=enable-chassis-as-gw
    ovs-vsctl set open . external_ids:ovn-bridge-mappings=${BRIDGE_MAPPINGS}

    # setup bridge
    GW_BRIDGE=`echo ${BRIDGE_PORTS} | cut -d":" -f1`
    GW_PORT=`echo ${BRIDGE_PORTS} | cut -d":" -f2`
    ovs-vsctl --may-exist add-br ${GW_BRIDGE}
    ovs-vsctl --may-exist add-port ${GW_BRIDGE} ${GW_PORT}
fi

ovs-appctl -t ovsdb-server ovsdb-server/add-remote ptcp:6640:127.0.0.1

tail -F --pid=$(cat /var/run/openvswitch/ovs-vswitchd.pid) /var/log/openvswitch/ovs-vswitchd.log &
tail -F --pid=$(cat /var/run/openvswitch/ovsdb-server.pid) /var/log/openvswitch/ovsdb-server.log &
wait

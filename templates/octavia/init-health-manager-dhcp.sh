#!/bin/bash
set -eux -o pipefail

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

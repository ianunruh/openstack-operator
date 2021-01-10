#!/bin/bash
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -ex

# Make the Nova Instances Dir as this is not autocreated.
mkdir -p /var/lib/nova/instances

# Set Ownership of nova dirs to the nova user
chown ${NOVA_USER_UID} /var/lib/nova /var/lib/nova/instances

# NOVA_HOME=/var/lib/nova
# mkdir -p ${NOVA_HOME}/.ssh
# cp /tmp/ssh-keys/id_ecdsa ${NOVA_HOME}/.ssh
# cp /tmp/ssh-keys/id_ecdsa.pub ${NOVA_HOME}/.ssh/authorized_keys
# cp /tmp/ssh-config/* /tmp/ssh-keys/ssh_host_* /etc/ssh
# chown -R ${NOVA_USER_UID} ${NOVA_HOME}/.ssh /etc/ssh

hypervisor_interface=$(ip -4 route list 0/0 | awk -F 'dev' '{ print $2; exit }' | awk '{ print $1 }') || exit 1
hypervisor_address=$(ip a s $hypervisor_interface | grep 'inet ' | awk '{print $2}' | awk -F "/" '{print $1}')

if [ -z "${hypervisor_address}" ] ; then
  echo "Var my_ip is empty"
  exit 1
fi

tee > /tmp/pod-shared/nova-hypervisor.conf << EOF
[DEFAULT]
my_ip  = $hypervisor_address
EOF

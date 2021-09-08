#!/bin/bash
set -ex

DEBIAN_FRONTEND=noninteractive dpkg-reconfigure openssh-server

NOVA_HOME=/var/lib/nova
mkdir -p ${NOVA_HOME}/.ssh
cp /tmp/ssh-keys/id_rsa ${NOVA_HOME}/.ssh
cp /tmp/ssh-keys/id_rsa.pub ${NOVA_HOME}/.ssh/authorized_keys

cat <<EOF > ${NOVA_HOME}/.ssh/config
Host *
    StrictHostKeyChecking no
    UserKnownHostsFile /dev/null
    Port 2022
    IdentitiesOnly yes
    SendEnv LANG LC_*
EOF

chown -R ${NOVA_USER_UID} ${NOVA_HOME}/.ssh

mkdir -p /run/sshd

exec /usr/sbin/sshd -D -e -p ${SSH_PORT}

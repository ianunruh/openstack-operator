#!/bin/bash
set -eux

cp /var/run/secrets/credential-keys/* /etc/keystone/credential-keys
chown -R keystone:keystone /etc/keystone/credential-keys
chmod 0440 /etc/keystone/credential-keys/*
chmod 0550 /etc/keystone/credential-keys

cp /var/run/secrets/fernet-keys/* /etc/keystone/fernet-keys
chown -R keystone:keystone /etc/keystone/fernet-keys
chmod 0440 /etc/keystone/fernet-keys/*
chmod 0550 /etc/keystone/fernet-keys

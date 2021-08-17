#!/bin/bash
set -ex

COMBINED_CERT=/etc/octavia/certs/client-combined/tls.crt

cp /etc/octavia/certs/client/tls.crt ${COMBINED_CERT}
chmod 0600 ${COMBINED_CERT}

# octavia expects certificate and private key to be combined
cat /etc/octavia/certs/client/tls.key >> ${COMBINED_CERT}

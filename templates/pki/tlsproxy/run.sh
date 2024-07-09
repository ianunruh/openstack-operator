#!/bin/bash
set -eu -o pipefail

cat /usr/local/etc/haproxy/certs/tls.crt > /usr/local/etc/haproxy/cert.pem
cat /usr/local/etc/haproxy/certs/tls.key >> /usr/local/etc/haproxy/cert.pem

exec haproxy -f /usr/local/etc/haproxy/haproxy.cfg

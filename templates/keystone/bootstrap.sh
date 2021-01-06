#!/bin/bash
set -e

keystone-manage --debug bootstrap \
    --bootstrap-password $KEYSTONE_ADMIN_PASSWORD \
    --bootstrap-admin-url $KEYSTONE_API_URL \
    --bootstrap-internal-url $KEYSTONE_API_INTERNAL_URL \
    --bootstrap-public-url $KEYSTONE_API_URL \
    --bootstrap-region-id $KEYSTONE_REGION

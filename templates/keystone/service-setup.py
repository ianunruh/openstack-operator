#!/usr/bin/env python3
import os

import openstack


def main():
    svc_name = os.environ['SVC_NAME']
    svc_type = os.environ['SVC_TYPE']
    svc_region = os.environ['SVC_REGION']

    endpoints = {
        'admin': os.environ['SVC_ENDPOINT_ADMIN'],
        'internal': os.environ['SVC_ENDPOINT_INTERNAL'],
        'public': os.environ['SVC_ENDPOINT_PUBLIC'],
    }

    openstack.enable_logging(debug=True)
    conn = openstack.connect()

    service = conn.identity.find_service(svc_name)
    if not service:
        service = conn.identity.create_service(name=svc_name, type=svc_type)

    region = conn.identity.find_region(svc_region)

    current_endpoints = conn.identity.endpoints(service_id=service.id)

    for endpoint_type, endpoint_url in endpoints.items():
        endpoint = next((x for x in current_endpoints if x.interface == endpoint_type and x.region_id == region.id), None)
        if endpoint:
            conn.identity.update_endpoint(endpoint, url=endpoint_url)
        else:
            conn.identity.create_endpoint(service_id=service.id, region_id=region.id, interface=endpoint_type, url=endpoint_url)


if __name__ == '__main__':
    main()

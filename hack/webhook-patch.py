#!/usr/bin/env python3
import json
import os
import sys
from urllib.parse import urlparse


def main():
    encoded = sys.stdin.read()

    config = json.loads(encoded)

    base_url = os.environ["WEBHOOK_BASE_URL"]

    annotations = config["metadata"]["annotations"]

    if "cert-manager.io/inject-ca-from" in annotations:
        del annotations["cert-manager.io/inject-ca-from"]

    for webhook in config["webhooks"]:
        clientConfig = webhook["clientConfig"]

        service = clientConfig.get("service")

        url = None

        if service:
            url = base_url + service["path"]
        else:
            parts = urlparse(clientConfig["url"])
            url = base_url + parts.path

        webhook["clientConfig"] = {
            "url": url,
        }

    print(json.dumps(config, indent=2))

main()

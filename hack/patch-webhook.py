#!/usr/bin/env python3
import json
import os
import sys


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

        if not service:
            continue

        webhook["clientConfig"] = {
            "url": base_url + service["path"],
        }

    print(json.dumps(config, indent=2))

main()

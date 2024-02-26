#!/usr/bin/env python
import os

import requests


class GitHubAPI:
    def __init__(self, username, token):
        self.username = username
        self.token = token

    def get_json(self, path, **kwargs):
        auth = (self.username, self.token)
        resp = requests.get(f"https://api.github.com/{path}", auth=auth, **kwargs)
        resp.raise_for_status()
        return resp.json()

    def delete(self, path, **kwargs):
        auth = (self.username, self.token)
        resp = requests.delete(f"https://api.github.com/{path}", auth=auth, **kwargs)
        resp.raise_for_status()


class GitHubRegistryAPI:
    def __init__(self):
        self.token = None

    def login(self, username, repo):
        params = {"scope": f"repository:{username}/{repo}:pull"}
        result = self.get_json("/token", params=params)
        self.token = result["token"]

    def get_json(self, path, **kwargs):
        headers = {}
        if self.token:
            headers["Authorization"] = f"Bearer {self.token}"
        resp = requests.get(f"https://ghcr.io/{path}", headers=headers, **kwargs)
        resp.raise_for_status()
        return resp.json()


def main():
    owner = "ianunruh"
    repo = "openstack-operator"

    max_versions_to_keep = 30

    ghcr = GitHubRegistryAPI()
    ghcr.login(owner, repo)

    github = GitHubAPI(owner, os.environ["GITHUB_TOKEN"])

    open_pulls = github.get_json(f"repos/{owner}/{repo}/pulls", params={"state": "open"})
    open_pull_tags = [f"pr-{pull['number']}" for pull in open_pulls]

    versions = github.get_json(f"user/packages/container/{repo}/versions", params={"per_page": 100})

    versions_to_keep = []
    for version in versions:
        version_tags = version["metadata"]["container"]["tags"]

        if "master" in version_tags:
            print(f"master - Keeping latest image - {version['name']}")
            continue

        if "dev" in version_tags:
            print(f"dev - Keeping latest image - {version['name']}")
            continue

        version_pull_tag = next((tag for tag in version_tags if tag in open_pull_tags), None)
        if version_pull_tag:
            print(f"{version_pull_tag} - Keeping latest image - {version['name']}")
            continue

        manifest = ghcr.get_json(f"v2/{owner}/{repo}/manifests/{version['name']}")
        image = ghcr.get_json(f"v2/{owner}/{repo}/blobs/{manifest['config']['digest']}")

        branch = image["config"]["Labels"]["org.opencontainers.image.version"]
        if branch == "master":
            if len(versions_to_keep) < max_versions_to_keep:
                print(f"master - Keeping historical image - {version['name']}")
                versions_to_keep.append(version)
                continue

        print(f"{branch} - Deleting historical image - {version['name']}")
        github.delete(f"user/packages/container/{repo}/versions/{version['id']}")


main()

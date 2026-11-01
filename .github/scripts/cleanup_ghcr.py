#!/usr/bin/env python3
# Copyright 2026 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import argparse
from dataclasses import dataclass
from datetime import datetime, timedelta, timezone
import json
import re
import subprocess
import sys
import time
from urllib.parse import quote

REPOSITORY = "chaos-mesh/chaos-mesh"
PACKAGE_NAMES = (
    "chaos-dlv",
    "chaos-daemon",
    "chaos-mesh",
    "chaos-dashboard",
    "chaos-kernel",
    "build-env",
    "dev-env",
)
INDEX_TYPES = {"application/vnd.oci.image.index.v1+json", "application/vnd.docker.distribution.manifest.list.v2+json"}
IMAGE_TYPES = {"application/vnd.oci.image.manifest.v1+json", "application/vnd.docker.distribution.manifest.v2+json"}


def require(condition, message):
    if not condition:
        raise ValueError(message)


def command(args):
    return subprocess.run(args, capture_output=True, text=True, timeout=180, check=False)


def api(path, method="GET"):
    result = command(["gh", "api", "--hostname", "github.com", "--include", "--method", method,
                      "-H", "Accept: application/vnd.github+json", "-H", "X-GitHub-Api-Version: 2022-11-28", path])
    header, separator, body = result.stdout.replace("\r\n", "\n").partition("\n\n")
    lines = header.splitlines()
    status = re.fullmatch(r"HTTP/\S+ (\d{3})(?: .*)?", lines[0] if lines else "")
    require(separator and status, f"Unreadable API response for {method} {path}")
    code = int(status[1])
    if result.returncode != 0 or code != (204 if method == "DELETE" else 200):
        detail = result.stderr.strip()
        try:
            error = json.loads(body)
        except json.JSONDecodeError:
            error = None
        if isinstance(error, dict) and isinstance(error.get("message"), str):
            detail = error["message"]
        raise ValueError(f"API {method} {path} returned HTTP {code} (gh exit {result.returncode}); "
                         f"{detail[:2000] or 'no error details'}; stopping without retry")
    return None if method == "DELETE" else json.loads(body)


def pages(path):
    records = []
    page = 1
    while True:
        separator = "&" if "?" in path else "?"
        data = api(f"{path}{separator}per_page=100&page={page}")
        require(isinstance(data, list), "Invalid API page")
        records.extend(data)
        if len(data) < 100:
            return records
        page += 1


def timestamp(value):
    return datetime.fromisoformat(value.replace("Z", "+00:00"))


@dataclass(frozen=True)
class Version:
    id: int
    digest: str
    tags: list
    created: datetime
    updated: datetime

    @classmethod
    def parse(cls, data):
        tags = data["metadata"]["container"]["tags"]
        require(isinstance(tags, list), "Invalid tags")
        return cls(data["id"], data["name"], tags, timestamp(data["created_at"]), timestamp(data["updated_at"]))

    def retained(self, cutoff):
        return bool(self.tags) or self.created >= cutoff or self.updated >= cutoff


def associated(package):
    repository = package["repository"]
    if repository is None:
        return False
    return repository["full_name"] == REPOSITORY


def references(image, value):
    result = command(["docker", "buildx", "imagetools", "inspect", "--raw", f"{image}@{value}"])
    require(result.returncode == 0, f"Cannot read manifest {image}@{value}")
    data = json.loads(result.stdout)
    media = data["mediaType"]
    require(media in INDEX_TYPES | IMAGE_TYPES, "Unsupported manifest media type")
    children = set()
    if media in INDEX_TYPES:
        require(isinstance(data.get("manifests"), list), "Missing manifest children")
        children.update(child["digest"] for child in data["manifests"])
    if "subject" in data:
        children.add(data["subject"]["digest"])
    return children


def cleanup(*, apply=False):
    if apply:
        cutoff = datetime.now(timezone.utc) - timedelta(days=30)
    for name in PACKAGE_NAMES:
        path = f"/orgs/chaos-mesh/packages/container/{quote(name, safe='')}"
        package = api(path)
        require(package["name"] == name, f"Unexpected package name for {name}")
        require(associated(package), f"Package {name} is not linked to {REPOSITORY}")
        versions = [Version.parse(data) for data in pages(f"{path}/versions?state=active")]
        tagged = sum(bool(version.tags) for version in versions)
        untagged = len(versions) - tagged
        if not apply:
            print(f"{name}: query verified: tagged={tagged}, untagged={untagged}", flush=True)
            continue
        print(f"{name}: before cleanup: tagged={tagged}, untagged={untagged}", flush=True)
        pending = {version.digest for version in versions if version.retained(cutoff)}
        candidates = [version for version in versions if version.digest not in pending]
        if not candidates:
            print(f"{name}: no old untagged versions; skipping", flush=True)
            continue
        protected = set()
        while pending:
            value = pending.pop()
            protected.add(value)
            pending.update(references(f"ghcr.io/chaos-mesh/{name}", value) - protected)
        candidates = [version for version in candidates if version.digest not in protected]
        deleted = 0
        try:
            for candidate in candidates:
                api(f"{path}/versions/{candidate.id}", method="DELETE")
                deleted += 1
                time.sleep(2)
        except Exception:
            print(f"{name}: cleanup stopped after {deleted} successful deletions", flush=True)
            raise
        print(f"{name}: after cleanup (calculated): tagged={tagged}, untagged={untagged - deleted}", flush=True)
    if not apply:
        print("Query verification complete; no image versions deleted.", flush=True)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Verify fixed GHCR package queries without deleting image versions.")
    parser.add_argument("--apply", action="store_true", help="delete eligible versions under the retention policy")
    args = parser.parse_args()
    try:
        cleanup(apply=args.apply)
    except (KeyError, TypeError, ValueError, OSError, subprocess.TimeoutExpired) as error:
        print(f"Cleanup stopped: {error}", file=sys.stderr, flush=True)
        sys.exit(1)

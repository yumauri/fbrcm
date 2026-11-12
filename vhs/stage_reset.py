#!/usr/bin/env python3
"""Stage one demo template as a historical fbrcm snapshot."""

import argparse
from datetime import datetime, timezone
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile

sys.dont_write_bytecode = True
from launch import PROJECTS, VHS


def cache_directory(binary, profile):
    env = {**os.environ, "FBRCM_OFFLINE": "1"}
    result = subprocess.run(
        [binary, "--profile", profile, "--no-local-config", "cache", "path", "--json"],
        env=env, check=True, stdout=subprocess.PIPE, text=True,
    )
    return Path(json.loads(result.stdout)["data"]["path"])


def template_content(template):
    result = {key: item for key, item in template.items() if key != "version"}
    for key, empty in (("parameters", {}), ("parameterGroups", {}), ("conditions", [])):
        if result.get(key) == empty:
            result.pop(key)
    return result


def stage(cache_dir, project, template):
    version = template["version"]["versionNumber"]
    if not isinstance(version, str) or not version.isascii() or not version.isdecimal() or int(version) < 1:
        raise ValueError("The fixture must have a positive Remote Config version number")
    cache_dir.mkdir(mode=0o700, parents=True, exist_ok=True)
    snapshot = cache_dir / f"{project}.{version}.json"
    payload = {
        "etag": "",
        "cached_at": datetime.now(timezone.utc).isoformat(timespec="seconds"),
        "remote_config": template,
    }
    # Link the completed file into place so readers never see partial JSON
    # and an existing immutable snapshot is never overwritten.
    with tempfile.TemporaryDirectory(prefix=".vhs-reset-", dir=cache_dir) as work:
        candidate = Path(work) / "snapshot.json"
        with open(candidate, "x", encoding="utf-8") as output:
            json.dump(payload, output, ensure_ascii=False)
        candidate.chmod(0o600)
        try:
            os.link(candidate, snapshot)
        except FileExistsError:
            existing = json.loads(snapshot.read_text(encoding="utf-8"))["remote_config"]
            if (existing["version"]["versionNumber"] != version
                    or template_content(existing) != template_content(template)):
                raise ValueError(f"Existing snapshot differs from the demo fixture: {snapshot}")
    return version


def main():
    fixtures = {project_id: filename for _, project_id, filename in PROJECTS}
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("project", choices=fixtures)
    parser.add_argument("--profile", default="personal")
    parser.add_argument("--binary", default="fbrcm")
    args = parser.parse_args()
    try:
        template = json.loads((VHS / fixtures[args.project]).read_text(encoding="utf-8"))
        version = stage(cache_directory(args.binary, args.profile), args.project, template)
    except (OSError, ValueError, KeyError, subprocess.CalledProcessError) as exc:
        parser.exit(1, f"Could not stage the demo snapshot: {exc}\n")
    print(f"Available in profile {args.profile}: client@{args.project} version {version}")


if __name__ == "__main__":
    main()

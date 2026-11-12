#!/usr/bin/env python3
"""Launch fbrcm with an isolated profile for demo recordings."""

import argparse
from datetime import datetime, timezone
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile


VHS = Path(__file__).resolve().parent
ROOT = VHS.parent
PROJECTS = (
    ("Mercato Mobile", "mercato-mobile-9eac5", "mercato_mobile.json"),
    ("Northstar Wallet", "northstar-wallet", "northstar_wallet.json"),
    ("PulseForge Fitness", "pulseforge-fitness-f60f7", "pulseforge_fitness.json"),
    ("SignalOps Console", "signalops-console", "signalops_console.json"),
    ("Streamline Play", "streamline-play-514406", "streamline_play.json"),
)


def write_json(path, data):
    path.write_text(json.dumps(data, indent=2) + "\n", encoding="utf-8")


def seed_live_auth(profile, cache_profile, source):
    if Path(source).name != source or source in (".", ".."):
        raise ValueError("Expected a profile name")

    home = Path.home()
    config_root = Path(os.environ.get("FBRCM_CONFIG_DIR") or
                       Path(os.environ.get("XDG_CONFIG_HOME") or home / ".config") / "fbrcm")
    default_cache = home / "Library" / "Caches" if sys.platform == "darwin" else Path(
        os.environ.get("XDG_CACHE_HOME") or home / ".cache")
    cache_root = Path(os.environ.get("FBRCM_CACHE_DIR") or default_cache / "fbrcm")
    auth = json.loads((config_root / "profiles" / source / "auth-config.json").read_text())
    entry = next(item for item in auth["auth"] if item["id"] == auth["default_auth_id"])
    token_path = entry.get("token_path")
    if not token_path:
        raise ValueError(f"Profile {source} needs a cached OAuth access token for live recording")
    token = json.loads((cache_root / "profiles" / source / token_path).read_text())
    expiry = datetime.fromisoformat(token["expiry"].replace("Z", "+00:00"))
    if not token.get("access_token") or (expiry - datetime.now(timezone.utc)).total_seconds() < 60:
        raise ValueError(f"Refresh the cached OAuth access token for profile {source} before recording")
    quota = entry.get("quota_project_id")
    if quota not in {project_id for _, project_id, _ in PROJECTS}:
        raise ValueError("The source profile's auth quota project must be a demo project")

    auth_path = Path("auth/demo")
    client = profile / auth_path / "client-secret.json"
    cached_token = cache_profile / auth_path / "token.json"
    client.parent.mkdir(parents=True)
    cached_token.parent.mkdir(parents=True)
    # fbrcm reads the OAuth descriptor before using its cached token. These
    # loopback endpoints and the absence of a refresh token keep it one-shot.
    write_json(client, {"installed": {
        "client_id": "vhs-recording",
        "auth_uri": "http://127.0.0.1:1/authorize",
        "token_uri": "http://127.0.0.1:1/token",
        "redirect_uris": ["http://localhost"],
    }})
    write_json(cached_token, {k: token[k] for k in ("access_token", "token_type", "expiry") if k in token})
    write_json(profile / "auth-config.json", {
        "version": 1, "default_auth_id": "demo", "auth": [{
            "id": "demo", "type": "oauth", "label": "Demo recording",
            "client_secret_path": (auth_path / "client-secret.json").as_posix(),
            "token_path": (auth_path / "token.json").as_posix(),
            "quota_project_id": quota,
        }],
    })
    return quota


def launch(args, session):
    config = session / "config"
    cache = session / "cache"
    profile = config / "profiles" / "demo"
    cache_profile = cache / "profiles" / "demo"
    templates = cache_profile / "remote-config"
    themes = config / "themes"
    for directory in (profile, templates, themes):
        directory.mkdir(parents=True)

    if args.theme != "default":
        theme = f"{args.theme}.toml"
        shutil.copyfile(ROOT / "themes" / theme, themes / theme)
        (config / "config.toml").write_text(f'theme = "{args.theme}"\n', encoding="utf-8")

    registry = {"version": 1, "projects": []}
    cached_at = datetime.now(timezone.utc).isoformat(timespec="seconds")
    for name, project_id, filename in PROJECTS:
        registry["projects"].append({"name": name, "project_id": project_id, "auth_id": "demo"})
        if args.live_profile:
            continue
        template = json.loads((VHS / filename).read_text(encoding="utf-8"))
        if args.empty_streamline and filename == "streamline_play.json":
            template = {"version": {}}
        write_json(templates / f"{project_id}.json", {
            "etag": f"demo-{project_id}",
            "cached_at": cached_at,
            "remote_config": template,
        })
    write_json(profile / "projects-config.json", registry)

    binary = args.binary or session / "fbrcm"
    if args.binary is None:
        subprocess.run(["go", "build", "-o", str(binary), "."], cwd=ROOT, check=True)

    env = os.environ.copy()
    env.pop("NO_COLOR", None)
    env.update({
        "FBRCM_CONFIG_DIR": str(config),
        "FBRCM_CACHE_DIR": str(cache),
        "FBRCM_NO_LOCAL_CONFIG": "1",
        "FBRCM_PROFILE": "demo",
        "FBRCM_LOG_LEVEL": "silent",
        "COLORTERM": "truecolor",
        "TZ": "UTC",
    })
    if args.live_profile:
        env["GOOGLE_CLOUD_QUOTA_PROJECT"] = seed_live_auth(profile, cache_profile, args.live_profile)
        env.pop("FBRCM_OFFLINE", None)
    else:
        env["FBRCM_OFFLINE"] = "1"
    return subprocess.call([str(binary)], cwd=session, env=env)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("theme", choices=("default", "fbrcm-dark", "fbrcm-light"))
    parser.add_argument("--binary", type=lambda path: Path(path).resolve())
    parser.add_argument("--live-profile", default=os.environ.get("FBRCM_VHS_LIVE_PROFILE"),
                        help="read from Firebase using this profile's cached OAuth access token")
    parser.add_argument("--empty-streamline", action="store_true")
    args = parser.parse_args()
    if args.live_profile and args.empty_streamline:
        parser.error("--empty-streamline requires the local fixtures")
    with tempfile.TemporaryDirectory(prefix="fbrcm-vhs-") as session:
        try:
            return launch(args, Path(session))
        except (ValueError, OSError, KeyError, StopIteration) as exc:
            parser.exit(1, f"Demo setup failed: {exc}\n")


if __name__ == "__main__":
    raise SystemExit(main())

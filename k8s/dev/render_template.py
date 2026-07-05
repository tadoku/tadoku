#!/usr/bin/env python3
import argparse
import json
from pathlib import Path


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--config", required=True)
    parser.add_argument("--environment", choices=["local", "shared"], required=True)
    parser.add_argument("--src", required=True)
    parser.add_argument("--dst", required=True)
    args = parser.parse_args()

    with open(args.config, "r", encoding="utf-8") as f:
        config = json.load(f)

    env_config = config[args.environment]
    hosts = env_config["hosts"]
    replacements = {
        "{{TADOKU_INGRESS_CLASS}}": env_config["ingress_class"],
        "{{TADOKU_APP_HOST}}": hosts["app"],
        "{{TADOKU_AUTH_HOST}}": hosts["auth"],
        "{{TADOKU_ADMIN_HOST}}": hosts["admin"],
        "{{TADOKU_APP_URL}}": "http://" + hosts["app"],
        "{{TADOKU_AUTH_URL}}": "http://" + hosts["auth"],
        "{{TADOKU_ADMIN_URL}}": "http://" + hosts["admin"],
        "{{TADOKU_COOKIE_DOMAIN}}": env_config["cookie_domain"],
    }

    src = Path(args.src)
    dst = Path(args.dst)
    content = src.read_text(encoding="utf-8")
    for needle, value in replacements.items():
        content = content.replace(needle, value)

    dst.parent.mkdir(parents=True, exist_ok=True)
    dst.write_text(content, encoding="utf-8")


if __name__ == "__main__":
    main()

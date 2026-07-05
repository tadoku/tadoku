#!/usr/bin/env python3
import argparse
import json
import sys
from pathlib import Path


def config_error(message: str) -> None:
    sys.exit(
        "invalid tilt config: {message}; "
        "copy tilt_config.json.example to tilt_config.json and fill it in".format(
            message=message
        )
    )


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--config", required=True)
    parser.add_argument("--environment", choices=["local", "shared"], required=True)
    parser.add_argument("--src", required=True)
    parser.add_argument("--dst", required=True)
    args = parser.parse_args()

    with open(args.config, "r", encoding="utf-8") as f:
        config = json.load(f)

    env_config = config.get(args.environment)
    if not isinstance(env_config, dict):
        config_error('missing "{env}" section'.format(env=args.environment))

    for key in ("ingress_class", "cookie_domain"):
        if key not in env_config:
            config_error('missing "{env}.{key}"'.format(env=args.environment, key=key))

    hosts = env_config.get("hosts")
    if not isinstance(hosts, dict):
        config_error('missing "{env}.hosts" section'.format(env=args.environment))
    for key in ("app", "auth", "admin"):
        if key not in hosts:
            config_error(
                'missing "{env}.hosts.{key}"'.format(env=args.environment, key=key)
            )

    ingress_class = env_config["ingress_class"]
    if ingress_class == "nginx":
        strip_path_suffix = "(/|$)(.*)"
        strip_path_type = "ImplementationSpecific"
    else:
        strip_path_suffix = ""
        strip_path_type = "Prefix"

    replacements = {
        "{{TADOKU_INGRESS_CLASS}}": ingress_class,
        "{{TADOKU_STRIP_PATH_SUFFIX}}": strip_path_suffix,
        "{{TADOKU_STRIP_PATH_TYPE}}": strip_path_type,
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

    if dst.exists() and dst.read_text(encoding="utf-8") == content:
        return
    dst.parent.mkdir(parents=True, exist_ok=True)
    dst.write_text(content, encoding="utf-8")


if __name__ == "__main__":
    main()

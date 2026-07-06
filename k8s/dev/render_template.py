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


def bool_config(env_config: dict, key: str, default: bool) -> bool:
    value = env_config.get(key, default)
    if not isinstance(value, bool):
        config_error('"{key}" must be true or false'.format(key=key))
    return value


def tls_block(host: str, secret_name: str) -> str:
    return (
        "  tls:\n"
        "    - hosts:\n"
        "        - {host}\n"
        "      secretName: {secret_name}\n"
    ).format(host=host, secret_name=secret_name)


def annotations_block(annotations: dict[str, str]) -> str:
    if not annotations:
        return ""
    lines = ["  annotations:"]
    for key, value in annotations.items():
        lines.append('    {key}: "{value}"'.format(key=key, value=value))
    return "\n".join(lines) + "\n"


def annotation_lines(annotations: dict[str, str]) -> str:
    if not annotations:
        return ""
    lines = []
    for key, value in annotations.items():
        lines.append('    {key}: "{value}"'.format(key=key, value=value))
    return "\n".join(lines) + "\n"


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

    default_scheme = "https" if args.environment == "shared" else "http"
    scheme = env_config.get("scheme", default_scheme)
    if scheme not in ("http", "https"):
        config_error(
            '"{env}.scheme" must be "http" or "https"'.format(env=args.environment)
        )

    tls_config = env_config.get("tls", {})
    if not isinstance(tls_config, dict):
        config_error('"{env}.tls" must be an object'.format(env=args.environment))
    tls_enabled = bool_config(
        tls_config,
        "enabled",
        args.environment == "shared" and scheme == "https",
    )
    tls_cluster_issuer = tls_config.get("cluster_issuer", "")
    if tls_enabled and not tls_cluster_issuer:
        config_error(
            'missing "{env}.tls.cluster_issuer"'.format(env=args.environment)
        )
    tls_secrets = tls_config.get("secret_names", {})
    if not isinstance(tls_secrets, dict):
        config_error(
            '"{env}.tls.secret_names" must be an object'.format(env=args.environment)
        )
    secret_names = {
        "app": tls_secrets.get("app", "tadoku-dev-app-tls"),
        "auth": tls_secrets.get("auth", "tadoku-dev-auth-tls"),
        "admin": tls_secrets.get("admin", "tadoku-dev-admin-tls"),
    }

    ssl_redirect = bool_config(env_config, "ssl_redirect", tls_enabled)
    if ssl_redirect and ingress_class != "nginx":
        config_error(
            '"{env}.ssl_redirect" requires ingress_class "nginx"; '
            'set it to false or use the nginx ingress class'.format(
                env=args.environment
            )
        )
    kratos_development = bool_config(
        env_config, "kratos_development", scheme != "https"
    )

    redirect_annotations = {}
    if ssl_redirect:
        redirect_annotations = {
            "nginx.ingress.kubernetes.io/ssl-redirect": "true",
            "nginx.ingress.kubernetes.io/force-ssl-redirect": "true",
        }

    cert_annotations = {}
    if tls_enabled:
        cert_annotations = {
            "cert-manager.io/cluster-issuer": tls_cluster_issuer,
        }

    replacements = {
        "{{TADOKU_INGRESS_CLASS}}": ingress_class,
        "{{TADOKU_STRIP_PATH_SUFFIX}}": strip_path_suffix,
        "{{TADOKU_STRIP_PATH_TYPE}}": strip_path_type,
        "{{TADOKU_APP_HOST}}": hosts["app"],
        "{{TADOKU_AUTH_HOST}}": hosts["auth"],
        "{{TADOKU_ADMIN_HOST}}": hosts["admin"],
        "{{TADOKU_APP_URL}}": scheme + "://" + hosts["app"],
        "{{TADOKU_AUTH_URL}}": scheme + "://" + hosts["auth"],
        "{{TADOKU_ADMIN_URL}}": scheme + "://" + hosts["admin"],
        "{{TADOKU_COOKIE_DOMAIN}}": env_config["cookie_domain"],
        "{{TADOKU_COOKIE_SECURE}}": str(scheme == "https").lower(),
        "{{TADOKU_KRATOS_DEVELOPMENT}}": str(kratos_development).lower(),
        "{{TADOKU_WEB_INGRESS_ANNOTATIONS}}": annotations_block(
            cert_annotations | redirect_annotations
        ),
        "{{TADOKU_API_INGRESS_ANNOTATIONS}}": annotations_block(
            redirect_annotations
        ),
        "{{TADOKU_AUTH_WEB_INGRESS_ANNOTATIONS}}": annotations_block(
            cert_annotations | redirect_annotations
        ),
        "{{TADOKU_ADMIN_WEB_INGRESS_ANNOTATIONS}}": annotations_block(
            cert_annotations | redirect_annotations
        ),
        "{{TADOKU_REDIRECT_ANNOTATION_LINES}}": annotation_lines(
            redirect_annotations
        ),
        "{{TADOKU_APP_TLS_BLOCK}}": tls_block(hosts["app"], secret_names["app"])
        if tls_enabled
        else "",
        "{{TADOKU_AUTH_TLS_BLOCK}}": tls_block(hosts["auth"], secret_names["auth"])
        if tls_enabled
        else "",
        "{{TADOKU_ADMIN_TLS_BLOCK}}": tls_block(hosts["admin"], secret_names["admin"])
        if tls_enabled
        else "",
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

#!/usr/bin/env python3
import argparse
import base64
import os
import re
import secrets
import subprocess
import sys
import tempfile
from pathlib import Path


TOKEN_PATTERN = re.compile(r"[0-9a-f]{64}")


def valid_token(value: str) -> bool:
    return TOKEN_PATTERN.fullmatch(value) is not None


def cluster_token(context: str) -> str | None:
    result = subprocess.run(
        [
            "kubectl",
            "--context",
            context,
            "--namespace",
            "default",
            "get",
            "secret",
            "dev-oathkeeper-authz",
            "--output=jsonpath={.data.token}",
        ],
        check=False,
        capture_output=True,
        text=True,
    )
    if result.returncode == 0:
        try:
            token = base64.b64decode(result.stdout, validate=True).decode("ascii")
        except (ValueError, UnicodeDecodeError) as exc:
            raise RuntimeError(
                "the existing dev-oathkeeper-authz Secret has an invalid token"
            ) from exc
        if not valid_token(token):
            raise RuntimeError(
                "the existing dev-oathkeeper-authz Secret token must be 64 lowercase hex characters"
            )
        return token

    if "NotFound" in result.stderr:
        return None

    detail = result.stderr.strip() or "kubectl returned no error details"
    raise RuntimeError("could not inspect the dev callback Secret: " + detail)


def cached_token(path: Path) -> str | None:
    try:
        token = path.read_text(encoding="ascii").strip()
    except FileNotFoundError:
        return None
    return token if valid_token(token) else None


def write_cache(path: Path, token: str) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    fd, temporary_name = tempfile.mkstemp(dir=path.parent, prefix=path.name + ".")
    temporary_path = Path(temporary_name)
    try:
        os.fchmod(fd, 0o600)
        with os.fdopen(fd, "w", encoding="ascii") as temporary_file:
            temporary_file.write(token + "\n")
        temporary_path.replace(path)
    finally:
        temporary_path.unlink(missing_ok=True)


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--context", required=True)
    parser.add_argument("--cache", required=True, type=Path)
    args = parser.parse_args()

    token = cluster_token(args.context)
    if token is None:
        token = cached_token(args.cache) or secrets.token_hex(32)
    write_cache(args.cache, token)
    print(token)


if __name__ == "__main__":
    try:
        main()
    except RuntimeError as exc:
        sys.exit(str(exc))

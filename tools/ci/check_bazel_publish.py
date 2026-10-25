#!/usr/bin/env python3
"""Exercise CI's real OCI runner against a disposable authenticated registry."""

import argparse
import base64
import json
import os
from pathlib import Path
import shutil
import subprocess
import time
import urllib.error
import urllib.request
import uuid


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--evidence-dir", type=Path, required=True)
    args = parser.parse_args()
    evidence = args.evidence_dir.resolve()
    evidence.mkdir(parents=True, exist_ok=False)
    root = Path(__file__).resolve().parents[2]
    workspace = evidence / "workspace"
    home = evidence / "home"
    workspace.mkdir()
    home.mkdir()
    name = "tadoku-bazel-auth-" + uuid.uuid4().hex[:12]
    image = "registry:2.8.3@sha256:a3d8aaa63ed8681a604f1dea0aa03f100d5895b6a58ace528858a7b332415373"
    # Test-only credential; this registry binds only to loopback.
    password = "tadoku-wrapper-test-only"
    (evidence / "htpasswd").write_text(
        "wrapper:$2y$05$abcdefghijklmnopqrstuuU4T8hlldLAroWJtHod80j2.hP98x7bO\n"
    )
    env = {k: v for k, v in os.environ.items() if k in (
        "PATH", "USER", "LANG", "LC_ALL", "SSL_CERT_FILE", "SSL_CERT_DIR",
        "BAZELISK_HOME", "BAZELISK_BASE_URL", "BAZELISK_FORMAT_URL",
        "DOCKER_HOST", "DOCKER_CONTEXT", "DOCKER_CERT_PATH", "DOCKER_TLS_VERIFY",
    )}
    env["HOME"] = str(home)
    commands = []

    def run(label, command, expected=0, stdin=None):
        commands.append({"label": label, "argv": command})
        with (evidence / f"{label}.log").open("w") as log:
            result = subprocess.run(command, cwd=workspace, env=env, input=stdin,
                                    text=True, stdout=log, stderr=subprocess.STDOUT, timeout=300)
        if (expected == 0 and result.returncode != 0) or (expected != 0 and result.returncode == 0):
            raise AssertionError(f"{label}: unexpected exit {result.returncode}; see {label}.log")
        return result.returncode

    bazel = [shutil.which("bazel"), "--ignore_all_rc_files", "--host_jvm_args=-Xmx512m", "--max_idle_secs=30",
             f"--output_user_root={evidence / 'bazel'}"]
    container_started = False
    result = {"revision": subprocess.check_output(
        ["git", "rev-parse", "HEAD"], cwd=root, text=True).strip(),
        "command": ["python3", str(Path(__file__).resolve()), "--evidence-dir", str(evidence)],
        "boundary": "Real Docker registry, rules_oci and CI image runner; synthetic credentials, loopback HTTP only. No GHCR publication."}
    try:
        run("registry-start", ["docker", "run", "--detach", "--name", name,
            "--memory=128m", "--cpus=1", "--publish", "127.0.0.1::5000",
            "--volume", f"{evidence / 'htpasswd'}:/auth/htpasswd:ro",
            "--env", "REGISTRY_AUTH=htpasswd", "--env", "REGISTRY_AUTH_HTPASSWD_REALM=wrapper-check",
            "--env", "REGISTRY_AUTH_HTPASSWD_PATH=/auth/htpasswd", image])
        container_started = True
        address = subprocess.check_output(
            ["docker", "port", name, "5000/tcp"], env=env, text=True).strip()
        result["registry_image"] = subprocess.check_output(
            ["docker", "inspect", "--format", "{{.Image}}", name], env=env, text=True).strip()
        for attempt in range(60):
            try:
                urllib.request.urlopen(f"http://{address}/v2/", timeout=1)
            except urllib.error.HTTPError as error:
                if error.code == 401:
                    break
                raise
            except OSError:
                time.sleep(0.25)
        else:
            raise AssertionError("registry did not require authentication")

        (workspace / "tools").mkdir()
        (workspace / "tools/bazel").symlink_to(root / "tools/bazel")
        shutil.copy(root / ".bazelversion", workspace)
        shutil.copy(root / "tools/ci/run_bazel_images.sh", workspace)
        (workspace / "MODULE.bazel").write_text('''module(name = "wrapper_publish_check")
bazel_dep(name = "rules_oci", version = "2.3.0")
bazel_dep(name = "rules_shell", version = "0.8.0")
''')
        (workspace / "BUILD.bazel").write_text('''load("@rules_oci//oci:defs.bzl", "oci_image", "oci_push")
load("@rules_shell//shell:sh_binary.bzl", "sh_binary")
oci_image(name = "image", architecture = "amd64", os = "linux")
oci_push(name = "push", image = ":image", repository = "%s/check/wrapper", remote_tags = ["check"])
sh_binary(
    name = "push_images",
    srcs = ["run_bazel_images.sh"],
    args = ["$(rlocationpath :push)"],
    data = [":push", "@bazel_tools//tools/bash/runfiles"],
)
''' % address)
        publish = bazel + ["run", "--jobs=1", "//:push_images"]
        run("unauthenticated-publish", publish, expected=1)
        failure = (evidence / "unauthenticated-publish.log").read_text()
        assert "UNAUTHORIZED" in failure.upper(), "failure must be registry authentication, not build setup"
        run("docker-login", ["docker", "login", address, "--username", "wrapper", "--password-stdin"],
            stdin=password + "\n")
        assert (home / ".docker/config.json").is_file()
        run("authenticated-publish", publish)
        request = urllib.request.Request(f"http://{address}/v2/check/wrapper/manifests/check",
            headers={"Authorization": "Basic " + base64.b64encode(f"wrapper:{password}".encode()).decode(),
                     "Accept": "application/vnd.oci.image.manifest.v1+json"})
        with urllib.request.urlopen(request, timeout=5) as response:
            assert response.status == 200
            result["published_digest"] = response.headers["Docker-Content-Digest"]
        config = evidence / "docker-config"
        shutil.move(str(home / ".docker"), config)
        env["DOCKER_CONFIG"] = str(config)
        run("docker-config-publish", publish)
        credential = base64.b64encode(f"wrapper:{password}".encode()).decode()
        for log in [*evidence.glob("*.log"), *(evidence / "bazel").glob("*/java.log*")]:
            contents = log.read_text(errors="replace")
            assert password not in contents and credential not in contents, f"credential leaked to {log.name}"
        result["result"] = "PASS: unauthenticated rejection; Docker login; HOME config and DOCKER_CONFIG publication; manifest readback"
        print(result["result"])
    except Exception as error:
        result["result"] = "FAIL: " + str(error)
        raise
    finally:
        if (workspace / "MODULE.bazel").exists():
            run("bazel-shutdown", bazel + ["shutdown"])
        if container_started:
            run("registry-remove", ["docker", "rm", "--force", name])
        for config in (home / ".docker/config.json", evidence / "docker-config/config.json"):
            config.unlink(missing_ok=True)
        result["commands"] = commands
        (evidence / "result.json").write_text(json.dumps(result, indent=2) + "\n")


if __name__ == "__main__":
    main()

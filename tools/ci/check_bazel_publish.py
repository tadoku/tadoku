#!/usr/bin/env python3
"""Exercise CI's real OCI runner against a disposable authenticated registry."""

import argparse
import base64
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import time
import urllib.error
import urllib.request
import uuid


FAILURE_CASES = [
    "unauthenticated push accepted or Docker config lost",
    "parallel publication remains serial or exceeds three children",
    "default image loading overlaps",
    "missing images, distinct digests or latest/prod tags",
    "failed child or unresolved runfile returns success or leaves children running",
]

# Only lifecycle failures use synthetic children; OCI publication and loading are real.
PROBE = r'''#!/usr/bin/env python3
import json, os, pathlib, resource, subprocess, sys, time
index, operation = sys.argv[1:]
record = {"image": int(index), "pid": os.getpid(), "start": time.monotonic()}
path = pathlib.Path(os.environ["PUBLISH_EVENTS"]) / (index + ".json")
path.write_text(json.dumps(record))
if operation == "lifecycle":
    time.sleep(0.05 if index == "1" else 0.4)
    status = 17 if index == "1" and os.environ.get("PUBLISH_FAIL_CHILD") else 0
else:
    command = [operation]
    if os.environ.get("PUBLISH_REPOSITORY"):
        command += ["--repository", os.environ["PUBLISH_REPOSITORY"] + "/image-" + index]
    child_env = dict(os.environ)
    if index == os.environ.get("PUBLISH_FAIL_IMAGE"):
        child_env["DOCKER_CONFIG"] = str(path.parent / "no-auth")
    child = subprocess.Popen(command, env=child_env)
    record["child_pid"] = child.pid
    status = child.wait()
usage = resource.getrusage(resource.RUSAGE_CHILDREN)
record.update(end=time.monotonic(), status=status, child_cpu_seconds=usage.ru_utime + usage.ru_stime,
              child_max_rss_kib=usage.ru_maxrss)
path.write_text(json.dumps(record))
sys.exit(status)
'''


def observations(directory):
    records = [json.loads(path.read_text()) for path in sorted(directory.glob("*.json"))]
    assert records, "no child executed"
    assert all("end" in record for record in records), "runner exited before children finished"
    for record in records:
        for key in ("pid", "child_pid"):
            assert key not in record or not Path(f"/proc/{record[key]}").exists(), "child not reaped"
    active = maximum = 0
    for _, change in sorted((record[key], change) for record in records
                            for key, change in (("start", 1), ("end", -1))):
        active += change
        maximum = max(maximum, active)
    return {"children": records, "max_active": maximum,
            "seconds": max(r["end"] for r in records) - min(r["start"] for r in records)}


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
        entry = {"label": label, "argv": command}
        commands.append(entry)
        start = time.monotonic()
        with (evidence / f"{label}.log").open("w") as log:
            result = subprocess.run(command, cwd=workspace, env=env, input=stdin,
                                    text=True, stdout=log, stderr=subprocess.STDOUT, timeout=300)
        entry.update(exit=result.returncode, seconds=time.monotonic() - start)
        if (expected == 0 and result.returncode != 0) or (expected != 0 and result.returncode == 0):
            raise AssertionError(f"{label}: unexpected exit {result.returncode}; see {label}.log")
        return result.returncode

    bazel = [shutil.which("bazel"), "--ignore_all_rc_files", "--host_jvm_args=-Xmx512m",
             "--host_jvm_args=-XX:ActiveProcessorCount=3", "--max_idle_secs=30",
             f"--output_user_root={evidence.parent / 'bazel-publish-output'}"]
    container_started = False
    result = {"revision": subprocess.check_output(
        ["git", "rev-parse", "HEAD"], cwd=root, text=True).strip(),
        "failure_cases": FAILURE_CASES,
        "command": ["python3", str(Path(__file__).resolve()), "--evidence-dir", str(evidence)],
        "boundary": "Real Docker registry, rules_oci and CI image runner; five config-only images, synthetic credentials, loopback HTTP only. No GHCR publication. Synthetic children cover lifecycle failures only."}
    (evidence / "tested.diff").write_bytes(subprocess.check_output(["git", "diff"], cwd=root))
    for source in (Path(__file__), root / "tools/ci/run_bazel_images.sh", root / "BUILD.bazel", root / "tools/bazel"):
        shutil.copyfile(source, evidence / (source.name + ".source"))
    load_tags = [f"{name}/image-{i}:check" for i in range(5)]
    try:
        run("registry-start", ["docker", "run", "--detach", "--name", name,
            "--memory=128m", "--cpus=1", "--pids-limit=64", "--label", "tadoku.fixture=" + name,
            "--read-only", "--tmpfs", "/var/lib/registry:rw,size=16m", "--publish", "127.0.0.1::5000",
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
        (workspace / "probe.py").write_text(PROBE)
        build = ['load("@rules_oci//oci:defs.bzl", "oci_image", "oci_push", "oci_load")',
                 'load("@rules_shell//shell:sh_binary.bzl", "sh_binary")']
        for i, tag in enumerate(load_tags):
            build += [f'oci_image(name = "image_{i}", architecture = "amd64", os = "linux", env = {{"CHECK_IMAGE": "{i}"}})',
                      f'oci_push(name = "push_{i}", image = ":image_{i}", repository = "{address}/check/default/image-{i}", remote_tags = ["latest", "prod"])',
                      f'oci_load(name = "load_{i}", image = ":image_{i}", repo_tags = ["{tag}"])']
            for operation in ("push", "load", "lifecycle"):
                target = f"{operation}_{i}"
                script = f"observe_{target}.sh"
                executable = f"push_{target}.sh" if operation == "push" else f"{target}.sh"
                command = '"lifecycle"' if operation == "lifecycle" else f'"$(rlocation _main/{executable})"'
                bootstrap = (root / "tools/ci/run_bazel_images.sh").read_text().split("resolve_runfile()")[0]
                (workspace / script).write_text(bootstrap + '\nexec python3 "$BUILD_WORKSPACE_DIRECTORY/probe.py" ' + str(i) + ' ' + command + '\n')
                (workspace / script).chmod(0o755)
                deps = ["@bazel_tools//tools/bash/runfiles"]
                if operation != "lifecycle":
                    deps.append(":" + target)
                build.append(f'sh_binary(name = "observe_{target}", srcs = ["{script}"], data = {deps!r})')
        for target, operation, parallel, count in (
                ("push_serial", "push", False, 5), ("push_parallel", "push", True, 5),
                ("load_images", "load", False, 5), ("child_failure", "lifecycle", True, 5),
                ("resolution_failure", "lifecycle", True, 2)):
            paths = [f"$(rlocationpath :observe_{operation}_{i})" for i in range(count)]
            if target == "resolution_failure":
                paths.append("missing-operation")
            deps = [f":observe_{operation}_{i}" for i in range(count)] + ["@bazel_tools//tools/bash/runfiles"]
            build.append(f'sh_binary(name = "{target}", srcs = ["run_bazel_images.sh"], args = {paths!r}, '
                         f'data = {deps!r}, env = {{"TADOKU_PARALLEL_IMAGE_PUSH": "{int(parallel)}"}})')
        (workspace / "BUILD.bazel").write_text("\n".join(build) + "\n")
        options = ["--jobs=3", "--local_resources=cpu=3", "--local_resources=memory=1536",
                   "--remote_upload_local_results=false"]
        # Build once before timing: each workload then uses the same native launch/runfiles path.
        run("build", bazel + ["build", *options, "//..."])

        def execute(label, target, repository=None, expected=0):
            events = evidence / (label + "-events")
            events.mkdir()
            env["PUBLISH_EVENTS"] = str(events)
            env["PUBLISH_FAIL_CHILD"] = "1" if label == "failed-child" else ""
            env["PUBLISH_FAIL_IMAGE"] = "1" if label == "failed-push" else ""
            env.pop("PUBLISH_REPOSITORY", None)
            if repository:
                env["PUBLISH_REPOSITORY"] = repository
            status = run(label, bazel + ["run", *options, "//:" + target], expected=expected)
            observed = observations(events)
            result[label] = observed | {"exit": status}
            return observed

        execute("unauthenticated-publish", "push_serial", expected=1)
        failure = (evidence / "unauthenticated-publish.log").read_text()
        assert "UNAUTHORIZED" in failure.upper(), "failure must be registry authentication, not build setup"
        run("docker-login", ["docker", "login", address, "--username", "wrapper", "--password-stdin"],
            stdin=password + "\n")
        assert (home / ".docker/config.json").is_file()

        def readback(repository, indices=range(5)):
            def fetch(path):
                request = urllib.request.Request(f"http://{address}/v2/{path}", headers={
                    "Authorization": "Basic " + base64.b64encode(f"wrapper:{password}".encode()).decode(),
                    "Accept": "application/vnd.oci.image.manifest.v1+json"})
                with urllib.request.urlopen(request, timeout=5) as response:
                    assert response.status == 200
                    return response.headers, response.read()
            images = []
            for i in indices:
                path = repository + f"/image-{i}"
                headers, manifest = fetch(path + "/manifests/latest")
                digest = "sha256:" + hashlib.sha256(manifest).hexdigest()
                assert headers["Docker-Content-Digest"] == digest
                assert fetch(path + "/manifests/prod")[1] == manifest
                assert fetch(path + "/manifests/" + digest)[1] == manifest
                config = json.loads(fetch(path + "/blobs/" + json.loads(manifest)["config"]["digest"])[1])
                assert f"CHECK_IMAGE={i}" in config["config"]["Env"]
                images.append({"repository": path, "digest": digest, "tags": ["latest", "prod"]})
            assert len({image["digest"] for image in images}) == len(images), "images are not distinct"
            return images

        for label, target in (("serial-publish", "push_serial"), ("parallel-publish", "push_parallel")):
            observed = execute(label, target, address + "/check/" + label)
            result[label]["images"] = readback("check/" + label)
            assert len(observed["children"]) == 5 and all(c["status"] == 0 for c in observed["children"])
            if label == "serial-publish":
                assert observed["max_active"] == 1, "default runner overlapped"
            else:
                assert 1 < observed["max_active"] <= 3, "parallel publication must overlap with at most three children"
        config = evidence / "docker-config"
        shutil.move(str(home / ".docker"), config)
        env["DOCKER_CONFIG"] = str(config)
        observed = execute("docker-config-publish", "push_parallel", address + "/check/docker-config")
        result["docker-config-publish"]["images"] = readback("check/docker-config")
        assert 1 < observed["max_active"] <= 3
        failed_push = execute("failed-push", "push_parallel", address + "/check/failed", expected=1)
        assert len(failed_push["children"]) == 5 and 1 < failed_push["max_active"] <= 3
        assert [c["image"] for c in failed_push["children"] if c["status"] != 0] == [1]
        assert "UNAUTHORIZED" in (evidence / "failed-push.log").read_text().upper()
        result["failed-push"]["images"] = readback("check/failed", (0, 2, 3, 4))
        loaded = execute("serial-load", "load_images")
        assert len(loaded["children"]) == 5 and loaded["max_active"] == 1, "default loads must remain serial"
        for i, tag in enumerate(load_tags):
            config = json.loads(subprocess.check_output(["docker", "image", "inspect", tag], env=env, text=True))[0]
            assert f"CHECK_IMAGE={i}" in config["Config"]["Env"]
        failed = execute("failed-child", "child_failure", expected=1)
        assert any(c["status"] == 17 for c in failed["children"]), "failure child never executed"
        assert 1 < failed["max_active"] <= 3
        assert len(failed["children"]) == 5
        unresolved = execute("unresolved-child", "resolution_failure", expected=1)
        assert len(unresolved["children"]) == 2 and all(c["status"] == 0 for c in unresolved["children"])
        credential = base64.b64encode(f"wrapper:{password}".encode()).decode()
        for log in [*evidence.glob("*.log"), *(evidence.parent / "bazel-publish-output").glob("*/java.log*")]:
            contents = log.read_text(errors="replace")
            assert password not in contents and credential not in contents, f"credential leaked to {log.name}"
        result["result"] = "PASS: authenticated bounded OCI publication; five distinct digests and both tags; serial Docker loads; failure propagation and reaping"
        print(result["result"])
    except Exception as error:
        result["result"] = "FAIL: " + str(error)
        raise
    finally:
        if (workspace / "MODULE.bazel").exists():
            run("bazel-shutdown", bazel + ["shutdown"])
        if container_started:
            run("registry-logs", ["docker", "logs", name])
            run("registry-stats", ["docker", "stats", "--no-stream", "--format", "{{json .}}", name])
            run("registry-remove", ["docker", "rm", "--force", "--volumes", name])
        for tag in load_tags:
            if subprocess.run(["docker", "image", "inspect", tag], env=env, capture_output=True).returncode == 0:
                run("remove-" + tag.split("/")[-1].replace(":", "-"), ["docker", "image", "rm", tag])
        result["cleanup"] = "owned registry, volumes and loaded tags removed; Bazel server shut down; Docker configs removed"
        for config in (home / ".docker/config.json", evidence / "docker-config/config.json"):
            config.unlink(missing_ok=True)
        result["commands"] = commands
        (evidence / "result.json").write_text(json.dumps(result, indent=2) + "\n")


if __name__ == "__main__":
    main()

#!/usr/bin/env python3
"""Run with Docker available; fixtures and registry stay on loopback."""

import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import time
import urllib.request


def run(*args, **kwargs):
    return subprocess.run(args, check=True, text=True, **kwargs)


script = Path(__file__).with_name("publish-image.sh").resolve()
scratch = Path(os.environ.get("TMPDIR", "/tmp"))
with tempfile.TemporaryDirectory(prefix="publish-image-", dir=scratch) as directory:
    root = Path(directory)
    frontend = root / "frontend"
    frontend.mkdir()
    (frontend / "scripts").mkdir()
    shutil.copy(script, frontend / "scripts/publish-image.sh")
    (frontend / "Dockerfile").write_text(
        "FROM alpine:3\nARG PROJECT_NAME\nCOPY . /fixture\n"
        "RUN test \"$PROJECT_NAME\" = webv2 && "
        "test ! -e /fixture/.env.local && "
        "grep -qx committed /fixture/value\n"
        "CMD [\"cat\", \"/fixture/value\"]\n"
    )
    (frontend / "value").write_text("committed\n")
    (root / ".gitignore").write_text(".env.local\n")
    run("git", "init", "-q", str(root))
    run("git", "-C", str(root), "add", ".")
    run("git", "-C", str(root), "-c", "user.name=Image check",
        "-c", "user.email=image-check@example.invalid", "commit", "-qm", "fixture")
    (frontend / ".env.local").write_text("SYNTHETIC_SECRET=must-not-enter-image\n")
    (frontend / "value").write_text("uncommitted\n")
    env = {**os.environ, "BUILD_WORKSPACE_DIRECTORY": str(root)}
    for args in [[], ["unknown"], ["webv2", "--tag", "t1"],
                 ["webv2", "--repository", "localhost:1/check"],
                 ["webv2", "--repository"],
                 ["webv2", "--repository", "localhost:1/check", "--tag", "bad/tag"],
                 ["webv2", "--repository", "https://bad", "--tag", "t1"]]:
        result = subprocess.run(["bash", str(script), *args], env=env,
                                capture_output=True, text=True)
        assert result.returncode != 0 and "Usage:" in result.stderr, result

    name = f"tadoku-publish-check-{os.getpid()}"
    run("docker", "run", "--rm", "-d", "--name", name,
        "-p", "127.0.0.1::5000", "registry:2", stdout=subprocess.DEVNULL)
    images = []
    try:
        address = run("docker", "port", name, "5000/tcp", capture_output=True).stdout.strip()
        registry = f"localhost:{address.rsplit(':', 1)[1]}"
        for attempt in range(30):
            try:
                urllib.request.urlopen(f"http://{registry}/v2/", timeout=1).close()
                break
            except OSError:
                time.sleep(1)
        else:
            raise AssertionError("Registry did not become ready")
        repository = f"{registry}/check/frontend-webv2"
        images = [f"{repository}:{tag}" for tag in ["t1", "t2"]]
        run("bash", str(script), "webv2", "--repository", repository,
            "--tag", "t1", "--tag", "t2", env=env, cwd=frontend)
        with urllib.request.urlopen(f"http://{registry}/v2/check/frontend-webv2/tags/list") as response:
            assert sorted(json.load(response)["tags"]) == ["t1", "t2"]
        run("docker", "image", "rm", *images, stdout=subprocess.DEVNULL)
        run("docker", "pull", images[0])
        assert run("docker", "run", "--rm", images[0], capture_output=True).stdout == "committed\n"
        print("PASS: rejected invalid inputs; committed tree only; requested tags only; registry pull and runtime.")
    finally:
        run("docker", "stop", name, stdout=subprocess.DEVNULL)
        if images:
            subprocess.run(["docker", "image", "rm", *images], capture_output=True)

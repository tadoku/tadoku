#!/usr/bin/env python3
"""Identify the native runtime used by the CI setup checks; probe errors disable reuse."""

import hashlib
import json
import os
from pathlib import Path
import platform
import shutil
import subprocess
import sys


def fingerprint():
    def command(*argv):
        return subprocess.check_output(argv, text=True, stderr=subprocess.STDOUT, timeout=20).strip()

    def digest(path):
        return hashlib.sha256(Path(path).read_bytes()).hexdigest()

    runtime = {
        "platform": [platform.system(), platform.machine(), platform.release(), platform.version()],
        "uid": os.getuid(),
        "environment": {key: os.environ.get(key) for key in (
            "PATH", "LANG", "LC_ALL", "LC_CTYPE", "TZ", "SSL_CERT_FILE", "SSL_CERT_DIR",
            "BAZELISK_BASE_URL", "BAZELISK_FORMAT_URL", "DOCKER_HOST", "DOCKER_CONTEXT",
            "DOCKER_TLS_VERIFY", "DOCKER_CERT_PATH",
            "USE_BAZEL_VERSION", "BAZELISK_SKIP_WRAPPER", "BAZEL_REAL",
            "GOENV", "GOFLAGS", "GOTOOLCHAIN", "GOROOT", "GOPATH", "GOWORK", "GODEBUG",
        )},
        "packages": command("dpkg-query", "-W", "-f=${Package}=${Version}\\n"),
        "python": sys.version,
        "tools": {},
        "rc": {},
    }
    for tool, args in {
        "bazel": ["--version"], "python3": ["--version"], "bash": ["--version"],
        "flock": ["--version"], "fuser": ["-V"], "git": ["--version"], "docker": ["--version"],
    }.items():
        path = shutil.which(tool)
        if path is None:
            raise RuntimeError(f"missing {tool}")
        runtime["tools"][tool] = [str(Path(path).resolve()), digest(path), command(tool, *args)]
    # Imports can name arbitrary external files; unfamiliar RC layouts rerun checks.
    for path in (Path('/etc/bazel.bazelrc'), Path.home() / '.bazelrc', Path('.bazelrc'),
                 Path('tools/oapi-codegen/.bazelrc'), Path.home() / '.bazeliskrc',
                 Path.home() / '.config/go/env'):
        if path.exists():
            if any(line.split()[:1] in (["import"], ["try-import"]) for line in path.read_text().splitlines()):
                raise RuntimeError(f"RC imports disable setup reuse: {path}")
            runtime['rc'][str(path)] = digest(path)
    runtime['docker_version'] = json.loads(command('docker', 'version', '--format', '{{json .}}'))
    info = json.loads(command('docker', 'info', '--format', '{{json .}}'))
    runtime['docker_info'] = {key: info[key] for key in (
        'Architecture', 'OSType', 'OperatingSystem', 'KernelVersion', 'ServerVersion', 'Driver',
        'CgroupDriver', 'CgroupVersion', 'SecurityOptions', 'Runtimes', 'DefaultRuntime', 'NCPU', 'MemTotal',
    )}
    runtime['limits'] = {}
    for name in ('cpu.max', 'memory.max', 'cpuset.cpus.effective'):
        path = Path('/sys/fs/cgroup') / name
        if path.exists():
            runtime['limits'][name] = path.read_text()
    return hashlib.sha256(json.dumps(runtime, sort_keys=True).encode()).hexdigest()


if __name__ == '__main__':
    try:
        print(fingerprint())
    except (OSError, ValueError, RuntimeError, KeyError, subprocess.SubprocessError) as error:
        sys.exit(f'Setup fingerprint unavailable; run checks: {error}')

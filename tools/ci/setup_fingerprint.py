#!/usr/bin/env python3
"""Identify the native runtime used by the CI setup checks; probe errors disable reuse."""

import argparse
import hashlib
import json
import os
from pathlib import Path
import platform
import shutil
import subprocess
import sys


WRAPPER_INPUTS = [
    "tools/bazel", ".bazelversion", "tools/ci/*.py", "tools/ci/run_bazel_images.sh",
    ".github/workflows/build-bazel.yaml",
]
OPENAPI_INPUTS = [
    "services/tadoku-api/spec/**", "services/tadoku-api/generated/openapi/**",
    "scripts/generate-openapi.sh", "tools/oapi-codegen/**", "tools/bazel", "tools/ci/*.py",
    ".github/workflows/build-bazel.yaml", ".bazel*", "go.mod", "go.sum", "**/MODULE.bazel*",
    "**/REPO.bazel", "**/BUILD*", "**/*.bzl", "**/*.patch",
]


def input_fingerprint(kind):
    patterns = WRAPPER_INPUTS if kind == 'wrapper' else OPENAPI_INPUTS
    root = Path(subprocess.check_output(['git', 'rev-parse', '--show-toplevel'], text=True).strip())
    subprocess.run(['git', 'diff', '--quiet', 'HEAD', '--', *patterns], cwd=root, check=True)
    unknown = subprocess.check_output(['git', 'ls-files', '--others', '-z', '--', *patterns], cwd=root)
    if unknown:
        raise RuntimeError('Untracked setup inputs disable reuse')
    manifest = subprocess.check_output(['git', 'ls-files', '--stage', '-z', '--', *patterns], cwd=root)
    records = [record.split(b'\t', 1) for record in manifest.split(b'\0') if record]
    if not records:
        raise RuntimeError('Empty setup input manifest')
    paths = {path for _, path in records}
    digest = hashlib.sha256()
    for metadata, path in records:
        source = root / os.fsdecode(path)
        if source.is_symlink() and os.fsencode(source.resolve(strict=True).relative_to(root)) not in paths:
            raise RuntimeError('Setup symlink target is outside the input closure')
        digest.update(metadata + b'\t' + path + b'\0' + str(source.lstat().st_mode).encode() + b'\0')
        digest.update(hashlib.sha256(source.read_bytes()).digest())
    return digest.hexdigest()


def resource_limits():
    root = Path('/sys/fs/cgroup')
    relative = next(line[3:] for line in Path('/proc/self/cgroup').read_text().splitlines() if line.startswith('0::'))
    scoped = root / relative.lstrip('/')
    if not scoped.is_dir():
        scoped = root
    return {name: (scoped / name).read_text() for name in ('cpu.max', 'memory.max', 'cpuset.cpus.effective')}


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
            "GO111MODULE", "GOOS", "GOARCH", "CGO_ENABLED", "XDG_CONFIG_HOME",
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
    if os.environ.get('GOWORK', '') not in ('', 'off'):
        raise RuntimeError('External Go workspaces disable setup reuse')
    goenv = os.environ.get('GOENV')
    config = Path(os.environ.get('XDG_CONFIG_HOME') or Path.home() / '.config')
    # Imports can name arbitrary external files; unfamiliar RC layouts rerun checks.
    for path in (Path('/etc/bazel.bazelrc'), Path.home() / '.bazelrc', Path('.bazelrc'),
                 Path('tools/oapi-codegen/.bazelrc'), Path.home() / '.bazeliskrc',
                 Path(goenv) if goenv and goenv != 'off' else config / 'go/env'):
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
    status = Path('/proc/self/status').read_text().splitlines()
    runtime['security'] = [line for line in status if line.split(':')[0] in (
        'Gid', 'Groups', 'CapInh', 'CapPrm', 'CapEff', 'CapBnd', 'CapAmb', 'NoNewPrivs', 'Seccomp',
    )]
    runtime['limits'] = resource_limits()
    return hashlib.sha256(json.dumps(runtime, sort_keys=True).encode()).hexdigest()


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--inputs', choices=('wrapper', 'openapi'))
    args = parser.parse_args()
    try:
        print(input_fingerprint(args.inputs) if args.inputs else fingerprint())
    except (OSError, ValueError, RuntimeError, KeyError, StopIteration, subprocess.SubprocessError) as error:
        sys.exit(f'Setup fingerprint unavailable; run checks: {error}')

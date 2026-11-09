#!/usr/bin/env python3
"""Prove the production config contract check passes on this tree and fails on a renamed or removed item."""

import json
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[2]
CHECK = HERE / 'check.py'


def run(root):
    return subprocess.run([sys.executable, str(CHECK), '--root', str(root)], capture_output=True, text=True)


def copy_tree(destination):
    contract = HERE / 'contract.json'
    files = {entry['file'] for entry in json.loads(contract.read_text())['entries']}
    for name in [str(contract.relative_to(ROOT)), *files]:
        target = destination / name
        target.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(ROOT / name, target)


def main():
    result = run(ROOT)
    assert result.returncode == 0, result.stdout + result.stderr
    checks = ['unchanged tree passes']

    with tempfile.TemporaryDirectory() as directory:
        root = Path(directory)
        copy_tree(root)
        assert run(root).returncode == 0
        urls = root / 'frontend/apps/webv3/src/lib/urls.server.ts'
        urls.write_text(urls.read_text().replace('TADOKU_API_URL', 'TADOKU_API_BASE_URL'))
        result = run(root)
        assert result.returncode == 1, result.stdout
        assert 'TADOKU_API_URL' in result.stderr and 'frontend/apps/webv3/src/lib/urls.server.ts' in result.stderr, \
            result.stderr
        checks.append('renamed env var fails and names it')

    with tempfile.TemporaryDirectory() as directory:
        root = Path(directory)
        copy_tree(root)
        flags = root / 'k8s/dev/base/flipt/features.yaml'
        flags.write_text(flags.read_text().replace('- key: release-webv3\n', '- key: release-webv3-preview\n'))
        result = run(root)
        assert result.returncode == 1 and 'release-webv3' in result.stderr, result.stderr
        checks.append('flag key renamed to a longer key fails')

    with tempfile.TemporaryDirectory() as directory:
        root = Path(directory)
        copy_tree(root)
        (root / 'frontend/apps/webv3/server/routes/healthz.ts').unlink()
        result = run(root)
        assert result.returncode == 1 and 'server/routes/healthz.ts' in result.stderr, result.stderr
        checks.append('removed probe route fails')

    print('PASS:', '; '.join(checks))


if __name__ == '__main__':
    main()

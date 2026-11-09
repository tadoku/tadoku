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


def run_mutated(name, file, old, new, appended):
    with tempfile.TemporaryDirectory() as directory:
        root = Path(directory)
        copy_tree(root)
        path = root / file
        text = path.read_text()
        assert old in text, f'{name}: {old!r} not in {file}'
        path.write_text(text.replace(old, new, 1) + appended)
        return run(root)


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

    dockerfile = 'frontend/apps/webv3/Dockerfile'
    urls = 'frontend/apps/webv3/src/lib/urls.server.ts'
    flags = 'k8s/dev/base/flipt/features.yaml'
    mutations = [
        ('prefixed env name', dockerfile, ' HOST=', ' APP_HOST='),
        ('prefixed env name', dockerfile, ' PORT=', ' APP_PORT='),
        ('extended host', dockerfile, 'TADOKU_WEB_URL=https://tadoku.app ', 'TADOKU_WEB_URL=https://tadoku.app.example '),
        ('prefixed host', urls, '`next.${web.host}`', '`preview-next.${web.host}`'),
        (
            'renamed env read kept in a comment',
            urls,
            'process.env.TADOKU_API_URL ',
            '/* formerly process.env.TADOKU_API_URL */ process.env.TADOKU_API_URL_V2 ',
        ),
        (
            'renamed segment kept in a comment',
            flags,
            '  - key: release-webv3-access\n',
            '  # - key: release-webv3-access\n  - key: release-webv3-beta\n',
        ),
        (
            'renamed env read kept in a line comment without a space',
            urls,
            'process.env.TADOKU_API_URL ',
            'process.env.TADOKU_API_URL_V2 ',
            '\nvoid 0;// formerly process.env.TADOKU_API_URL\n',
        ),
        (
            'renamed Go flag key kept in a line comment without a space',
            'services/tadoku-api/internal/featureflags/registry.go',
            '"release-webv3",',
            '"release-webv3-next",',
            '\nvar _ = 0// formerly "release-webv3"\n',
        ),
        ('extended host template', urls, '`next.${web.host}`', '`next.${web.host}.example`'),
        ('env name extended with $', urls, 'process.env.TADOKU_API_URL ', 'process.env.TADOKU_API_URL$ '),
    ]
    harmless = [
        ('block comment opener in a string', urls, '  const api = ', "  void '/*'\n  const api = "),
        ('line comment opener in a string', urls, '  const api = ', "  void ' // '; const api = "),
        ('hash in a Dockerfile value', dockerfile, 'ENV NODE_ENV=', 'ENV NOTE=" # " NODE_ENV='),
    ]
    for name, file, old, new, *appended in mutations:
        result = run_mutated(name, file, old, new, ''.join(appended))
        assert result.returncode == 1 and file in result.stderr, f'{name} passed: {result.stdout}'
        checks.append(f'{name} fails')
    for name, file, old, new in harmless:
        result = run_mutated(name, file, old, new, '')
        assert result.returncode == 0, f'{name} failed: {result.stderr}'
        checks.append(f'{name} passes')

    print('PASS:', '; '.join(checks))


if __name__ == '__main__':
    main()

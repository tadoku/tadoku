#!/usr/bin/env python3
"""Prove setup keys cover names, modes, bytes, symlinks and clean-checkout boundaries."""

import argparse
import json
from pathlib import Path
import subprocess
import sys
import tempfile


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--evidence-dir', required=True, type=Path)
    evidence = parser.parse_args().evidence_dir.resolve()
    evidence.mkdir(parents=True, exist_ok=False)
    script = Path(__file__).with_name('setup_fingerprint.py').resolve()
    checks = []
    with tempfile.TemporaryDirectory(dir=evidence) as directory:
        root = Path(directory)

        def git(*args):
            return subprocess.check_output(['git', '-c', 'core.hooksPath=/dev/null', *args], cwd=root)

        def commit():
            git('add', '--all')
            git('-c', 'user.name=Setup check', '-c', 'user.email=setup@example.invalid', 'commit', '-qm', 'Fixture')

        def key(kind, expected=0):
            result = subprocess.run([sys.executable, str(script), '--inputs', kind], cwd=root,
                                    capture_output=True, text=True)
            assert result.returncode == expected, result.stderr
            if expected:
                assert not result.stdout
            return result.stdout.strip()

        git('init', '-q')
        for name, contents in {'.bazelversion': '9.2.0\n', 'tools/bazel': '#!/bin/sh\n',
                               'services/tadoku-api/spec/openapi/root.yaml': 'paths: {}\n'}.items():
            path = root / name
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_text(contents)
        commit()
        wrapper, openapi = key('wrapper'), key('openapi')
        assert wrapper == key('wrapper') and openapi == key('openapi')
        checks.append('unchanged keys stable')
        (root / 'notes.md').write_text('unrelated documentation\n')
        commit()
        assert wrapper == key('wrapper') and openapi == key('openapi')
        checks.append('unrelated documentation does not invalidate')
        fragment = root / 'services/tadoku-api/spec/openapi/root.yaml'
        renamed = fragment.with_name('renamed.yaml')
        fragment.rename(renamed)
        commit()
        assert openapi != key('openapi') and wrapper == key('wrapper')
        checks.append('same-byte fragment rename invalidates')
        openapi = key('openapi')
        renamed.unlink()
        commit()
        assert openapi != key('openapi')
        checks.append('fragment deletion invalidates')
        hook = root / 'tools/bazel'
        hook.write_text('#!/bin/sh\nexit 0\n')
        commit()
        assert wrapper != key('wrapper')
        checks.append('committed input bytes invalidate')
        wrapper = key('wrapper')
        hook.chmod(0o755)
        commit()
        assert wrapper != key('wrapper')
        checks.append('executable mode invalidates')
        hook.write_text('#!/bin/sh\nexit 1\n')
        key('wrapper', expected=1)
        checks.append('dirty tracked input disables reuse')
        git('restore', 'tools/bazel')
        fragment.write_text('paths: {}\n')
        key('openapi', expected=1)
        checks.append('untracked fragment disables reuse')
        commit()
        (root / '.gitignore').write_text('*.out\n')
        commit()
        ignored = fragment.with_name('unexpected.out')
        ignored.write_text('paths: {}\n')
        key('openapi', expected=1)
        checks.append('ignored untracked fragment disables reuse')
        ignored.unlink()
        link = root / 'tools/oapi-codegen/tools/bazel'
        link.parent.mkdir(parents=True)
        link.symlink_to('../../bazel')
        commit()
        key('openapi')
        checks.append('tracked in-closure symlink accepted')
        link.unlink()
        link.symlink_to(evidence / 'outside')
        (evidence / 'outside').write_text('outside fixture\n')
        commit()
        key('openapi', expected=1)
        checks.append('out-of-closure symlink disables reuse')
    (evidence / 'result.json').write_text(json.dumps({'passed': checks}, indent=2) + '\n')
    print('PASS:', '; '.join(checks))


if __name__ == '__main__':
    main()

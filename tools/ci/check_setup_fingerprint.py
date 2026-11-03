#!/usr/bin/env python3
"""Exercise native setup fingerprint stability and fail-open invalidation."""

import argparse
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--evidence-dir', type=Path, required=True)
    evidence = parser.parse_args().evidence_dir.resolve()
    evidence.mkdir(parents=True, exist_ok=False)
    script = Path(__file__).with_name('setup_fingerprint.py').resolve()
    observations = []

    def run(label, extra=None, expected=0):
        result = subprocess.run([sys.executable, str(script)], env=dict(os.environ) | (extra or {}),
                                capture_output=True, text=True)
        observations.append({'case': label, 'exit': result.returncode, 'output': result.stdout.strip()})
        assert result.returncode == expected, result.stderr
        if expected:
            assert not result.stdout, 'probe failure emitted a usable fingerprint'
        return result.stdout

    baseline = run('baseline')
    assert len(baseline.strip()) == 64
    assert baseline == run('unchanged')
    assert baseline != run('locale-change', {'LC_ALL': 'C' if os.environ.get('LC_ALL') != 'C' else 'C.UTF-8'})
    with tempfile.TemporaryDirectory(dir=evidence) as directory:
        env = {'HOME': directory}
        before = run('isolated-home', env)
        rc = Path(directory) / '.bazelrc'
        rc.write_text('build --jobs=1\n')
        assert before != run('rc-change', env)
        rc.write_text('import /does/not/exist\n')
        run('unknown-rc-input', env, expected=1)
    run('missing-tools', {'PATH': '/does/not/exist'}, expected=1)
    (evidence / 'result.json').write_text(json.dumps({'passed': observations}, indent=2) + '\n')
    print('PASS: native stability, locale/RC invalidation, unavailable input/tool fallback')


if __name__ == '__main__':
    main()

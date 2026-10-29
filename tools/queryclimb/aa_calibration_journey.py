#!/usr/bin/env python3
import copy
import json
import pathlib
import subprocess
import sys

binary, evidence = sys.argv[1:]
out = pathlib.Path(evidence)
out.mkdir(parents=True, exist_ok=True)

# Failure modes: missing or A/B-disguised calibration; wrong/missing/duplicate
# cases or scales; malformed raw samples; high A/A noise ignored in acceptance.
query = 'target.Read'
metric = dict(query=query, case='qualified', fingerprint='same', rows_processed=10,
              shared_buffers=10, temp_buffers=0)
ab = dict(query=query, case='qualified', baseline_ms=[100] * 15, candidate_ms=[90] * 15,
          aa_control=False, probe_ms=[[100] * 5, [100] * 5], batch_noise=[0] * 30)
aa = dict(ab, candidate_ms=[100] * 15, aa_control=True)
checks = []

def run(name, calibration=((aa,),), timing=ab, expected=False):
    folder = out / name
    folder.mkdir()
    for filename, value in [('metrics.json', [metric]), ('latency.json', [timing])]:
        (folder / filename).write_text(json.dumps(value))
    command = [binary, 'gate', '--mode', 'candidate', '--target', query,
               '--baseline', str(folder / 'metrics.json'), '--candidate', str(folder / 'metrics.json'),
               '--latency', str(folder / 'latency.json'), '--output', str(folder / 'gate.json')]
    if calibration is not None:
        files = []
        for scale, samples in enumerate(calibration):
            path = folder / ('aa-' + str(scale) + '.json')
            path.write_text(json.dumps(samples))
            files.append(str(path))
        command += ['--aa-calibration', ','.join(files)]
    result = subprocess.run(command, capture_output=True, text=True)
    (folder / 'command.json').write_text(json.dumps(command))
    (folder / 'output.log').write_text(result.stdout + result.stderr)
    checks.append(dict(name=name, exit=result.returncode, expected=expected))
    (out / 'checks.json').write_text(json.dumps(checks, indent=2))
    assert (result.returncode == 0) == expected, (name, result.stdout, result.stderr)

run('missing-aa', calibration=None)
run('matched-raw-aa', expected=True)
for field, value in [('aa_control', False), ('query', 'other.Read'), ('case', 'other'),
                     ('baseline_ms', [100] * 14), ('candidate_ms', [100] * 14),
                     ('candidate_ms', [112] * 15), ('candidate_ms', [0] * 15),
                     ('probe_ms', [[100] * 4, [100] * 5]), ('batch_noise', []),
                     ('batch_noise', [-1] * 30), ('probe_ms', [[100] * 5, [112] * 5]),
                     ('batch_noise', [0.11] * 30)]:
    changed = copy.deepcopy(aa)
    changed[field] = value
    run('invalid-' + field + '-' + str(len(checks)), calibration=((changed,),))
run('missing-case', calibration=((),))
run('duplicate-case', calibration=((aa, aa),))
run('extra-case', calibration=((aa, dict(aa, case='extra')),))
run('wrong-scale', calibration=((), (aa,)))
run('extra-empty-scale', calibration=((aa,), ()))
run('malformed-json-shape', calibration=({'invalid': 1},))
# Stored aggregate claims cannot hide actual A/A noise. A 5% gain clears the
# A/B-only 2% floor, but cannot clear the 12% threshold from 6% A/A noise.
noisy_aa = dict(aa, candidate_ms=[106] * 15, noise=0, threshold=0.02)
run('aa-noise-raises-improvement-threshold', calibration=((noisy_aa,),),
    timing=dict(ab, candidate_ms=[95] * 15))
run('improvement-clears-aa-threshold', calibration=((noisy_aa,),),
    timing=dict(ab, candidate_ms=[85] * 15), expected=True)
# The A/B noise still wins when greater than A/A noise.
run('ab-noise-still-applies', timing=dict(ab, candidate_ms=[95] * 15, batch_noise=[0.06] * 30))
(out / 'summary.json').write_text(json.dumps(dict(status='passed', binary=binary, checks=len(checks),
    boundary='Real CLI with synthetic raw latency samples; no timed batches executed.'), indent=2))
print('PASS: required raw A/A calibration, scope and noise thresholds')

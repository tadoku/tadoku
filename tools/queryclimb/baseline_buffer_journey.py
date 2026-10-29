#!/usr/bin/env python3
import copy
import hashlib
import json
import pathlib
import subprocess
import sys

binary, evidence = sys.argv[1:]
out = pathlib.Path(evidence)
out.mkdir(parents=True, exist_ok=True)

# Failure modes: candidate SQL or gate mode gets the baseline exception; scope,
# identity, full plans, fingerprints, work, shared buffers, counters or inputs
# change; witnesses are missing, repeated, planless or have no temp variance.
def digest(value):
    return hashlib.sha256(value.encode()).hexdigest()

query, case = 'logs.ListLogsForContest', 'past-end-all'
key = '0/' + query + '/' + case
snap = {part: digest(part) for part in ['schema', 'data', 'statistics', 'settings', 'migrations']}
a = dict(query=query, case=case, database='qc_baseline', sql_hash=digest('baseline SQL'),
         fingerprint=digest('empty page'), query_id=17, rows_read=250, rows_processed=500,
         shared_buffers=100, temp_buffers=40, plan_hash=digest('shape'),
         generic_plans=1, custom_plans=5, returned=0, affected=0,
         tenant='tadoku/prod', args=['2147483647', '25'], warm_args=[['0', '25']] * 5,
         tie_columns=[0], mask_columns=[], indexes=['logs_pkey'], signatures=['sort-spill'],
         estimated_plan=[{'Plan': {'Node Type': 'Sort', 'Plan Rows': 250,
                                  'Total Cost': 800, 'Plans': [{'Node Type': 'Seq Scan',
                                                            'Relation Name': 'logs'}]}}])
b = dict(a, database='qc_repeat', temp_buffers=38)
witness = [copy.deepcopy(a), copy.deepcopy(b), dict(a, database='qc_third')]
proof = dict(target=query, target_read_only=True, baseline_database=a['database'],
             candidate_database=b['database'], baseline=copy.deepcopy(snap), candidate=copy.deepcopy(snap),
             queries={}, variance={key: witness},
             baseline_buffer=dict(key=key, sql_hash=a['sql_hash'],
                                  identities={m['database']: copy.deepcopy(snap) for m in witness}))
checks = []

def run(name, original=a, candidate=b, identity=proof, mode='determinism', expected=False):
    folder = out / name
    folder.mkdir()
    for filename, value in [('a.json', [original]), ('b.json', [candidate]), ('proof.json', identity)]:
        (folder / filename).write_text(json.dumps(value))
    command = [binary, 'gate', '--mode', mode, '--target', query,
               '--baseline', str(folder / 'a.json'), '--candidate', str(folder / 'b.json'),
               '--buffer-proof', str(folder / 'proof.json'), '--output', str(folder / 'gate.json')]
    if mode == 'candidate':
        latency = dict(query=query, case=case, baseline_ms=[100] * 15, candidate_ms=[90] * 15,
                       probe_ms=[[100] * 5, [100] * 5], batch_noise=[0] * 30, aa_control=False)
        (folder / 'latency.json').write_text(json.dumps([latency]))
        (folder / 'aa.json').write_text(json.dumps([dict(latency, aa_control=True, candidate_ms=[100] * 15)]))
        command += ['--latency', str(folder / 'latency.json'), '--aa-calibration', str(folder / 'aa.json')]
    result = subprocess.run(command, capture_output=True, text=True)
    (folder / 'command.json').write_text(json.dumps(command))
    (folder / 'output.log').write_text(result.stdout + result.stderr)
    checks.append(dict(name=name, exit=result.returncode, expected=expected))
    (out / 'checks.json').write_text(json.dumps(checks, indent=2))
    assert (result.returncode == 0) == expected, (name, result.stdout, result.stderr)
    if expected:
        gate = json.loads((folder / 'gate.json').read_text())
        assert gate['pass'] and len(gate['assertion_replacements']) == 1, gate

run('baseline-temp-only', expected=True)
for mode in ['guard', 'candidate']:
    run('wrong-mode-' + mode, mode=mode)
for field, value in [('sql_hash', digest('candidate SQL')), ('case', 'different'),
                     ('fingerprint', digest('different rows')), ('rows_read', 251),
                     ('rows_processed', 501), ('shared_buffers', 101),
                     ('plan_hash', digest('different shape')), ('generic_plans', 2),
                     ('custom_plans', 4), ('args', ['0', '25']), ('warm_args', []),
                     ('tenant', 'test/other'), ('tie_columns', [1]), ('mask_columns', [1]),
                     ('returned', 1), ('affected', 1), ('indexes', []), ('signatures', []),
                     ('estimated_plan', None),
                     ('estimated_plan', [{'Plan': {'Node Type': 'Sort', 'Plan Rows': 251}}])]:
    changed = copy.deepcopy(b)
    changed[field] = value
    run('candidate-' + field + '-' + str(len(checks)), candidate=changed)
    changed_proof = copy.deepcopy(proof)
    changed_proof['variance'][key][2][field] = value
    run('witness-' + field + '-' + str(len(checks)), identity=changed_proof)

for scope in ['1/' + query + '/' + case, '0/' + query + '/other', '0/other.Read/' + case]:
    changed = copy.deepcopy(proof)
    changed['baseline_buffer']['key'] = scope
    run('wrong-scope-' + str(len(checks)), identity=changed)
for name in ['baseline', 'candidate']:
    for part in snap:
        for missing in [False, True]:
            changed = copy.deepcopy(proof)
            if missing:
                del changed[name][part]
            else:
                changed[name][part] = digest('changed')
            run(name + '-' + part + '-' + str(missing), identity=changed)
for db in proof['baseline_buffer']['identities']:
    for part in snap:
        for missing in [False, True]:
            changed = copy.deepcopy(proof)
            if missing:
                del changed['baseline_buffer']['identities'][db][part]
            else:
                changed['baseline_buffer']['identities'][db][part] = digest('changed')
            run(db + '-' + part + '-' + str(missing), identity=changed)
for name in ['missing-certificate', 'missing-sql', 'changed-sql', 'missing-identities',
             'missing-clone', 'extra-clone', 'two-witnesses', 'four-witnesses',
             'duplicate-clone', 'no-temp-variance', 'shared-only-variance']:
    changed = copy.deepcopy(proof)
    cert = changed['baseline_buffer']
    samples = changed['variance'][key]
    if name == 'missing-certificate':
        del changed['baseline_buffer']
    elif name == 'missing-sql':
        del cert['sql_hash']
    elif name == 'changed-sql':
        cert['sql_hash'] = digest('candidate SQL')
    elif name == 'missing-identities':
        del cert['identities']
    elif name == 'missing-clone':
        del cert['identities']['qc_third']
    elif name == 'extra-clone':
        cert['identities']['qc_extra'] = copy.deepcopy(snap)
    elif name == 'two-witnesses':
        samples.pop()
    elif name == 'four-witnesses':
        samples.append(dict(a, database='qc_fourth'))
    elif name == 'duplicate-clone':
        samples[2]['database'] = samples[0]['database']
    elif name in ['no-temp-variance', 'shared-only-variance']:
        for sample in samples:
            sample['temp_buffers'] = a['temp_buffers']
        if name == 'shared-only-variance':
            samples[2]['shared_buffers'] += 1
    run(name, identity=changed)

changed = copy.deepcopy(b)
changed['shared_buffers'] += 1
run('shared-and-temp-variance', candidate=changed)
(out / 'summary.json').write_text(json.dumps(dict(status='passed', binary=binary, checks=len(checks),
    boundary='Real CLI with synthetic artifacts; live collector and retained-clone evidence checked separately.'), indent=2))
print('PASS: baseline-only temporary-buffer certification and adversarial counterexamples')

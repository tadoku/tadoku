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

# Failure modes: illegal tie member/value, wrong boundary cardinality, lost NULL,
# lost interior row, duplicate row, reversed groups, target-work regression,
# unchanged-query exemption with changed SQL/schema/data/stats/settings/plan,
# an exemption without observed variance, or silently weakening the strict gate.
def digest(value):
    return hashlib.sha256(value.encode()).hexdigest()

def encoded(value):
    return json.dumps(value, separators=(',', ':'), ensure_ascii=False)

def fingerprint(rows):
    groups = []
    for row in rows:
        if not groups or groups[-1][0] != row[1]:
            groups.append((row[1], []))
        groups[-1][1].append(encoded(row))
    return digest(encoded([row for _, group in groups for row in sorted(group)]))

universe = [['a', '9', 'A'], ['b', '9', None], ['c', '8', 'C'], ['d', '8', 'D'],
            ['e', '8', 'E'], ['f', '7', 'F'], ['g', '7', 'G'], ['h', '6', 'H']]
before = [universe[2], universe[3], universe[5]]
after = [universe[4], universe[3], universe[6]]

def metric(query, rows, sql):
    return dict(query=query, case='boundary', query_id=1, rows_processed=30,
                shared_buffers=71, temp_buffers=3, plan_hash=digest('shape'),
                fingerprint=fingerprint(rows), rows=rows, args=['3', '3'],
                tenant='tadoku/prod', tie_columns=[1], sql_hash=digest(sql),
                generic_plans=1, custom_plans=5, returned=len(rows),
                estimated_plan=[{'Plan': {'Node Type': 'Sort', 'Plan Rows': 25,
                                         'Total Cost': 10, 'Workers Planned': 1}}])

a = metric('target.Read', before, 'baseline')
b = metric('target.Read', after, 'candidate')
other = metric('other.Read', [['x', '1', None]], 'unchanged')
other_after = copy.deepcopy(other)
other_after['shared_buffers'] = 72
other_after['temp_buffers'] = 4
oracle = dict(query=a['query'], case=a['case'], tenant=a['tenant'], args=a['args'],
              tie_columns=[1], offset_arg=0, limit_arg=1,
              sql_hash=a['sql_hash'], rows=universe)
observations = []
for shared, temp in [(71, 3), (72, 4), (71, 3)]:
    sample = copy.deepcopy(other)
    sample.update(shared_buffers=shared, temp_buffers=temp)
    observations.append(sample)
snap = {key: digest(key) for key in ['schema', 'data', 'statistics', 'settings', 'migrations']}
proof = dict(target='target.Read', target_read_only=True, baseline=snap,
             candidate=copy.deepcopy(snap), queries={'other.Read': other['sql_hash']},
             variance={'0/other.Read/boundary': observations})

def run(name, target=b, unchanged=other_after, page=oracle, identity=proof,
        expected=True, amended=True):
    folder = out / name
    folder.mkdir()
    for filename, value in [('a.json', [a, other]), ('b.json', [target, unchanged]),
                            ('pages.json', [page]), ('proof.json', identity)]:
        (folder / filename).write_text(json.dumps(value))
    cmd = [binary, 'gate', '--mode', 'guard', '--target', 'target.Read',
           '--baseline', str(folder / 'a.json'), '--candidate', str(folder / 'b.json'),
           '--output', str(folder / 'gate.json')]
    if amended:
        cmd += ['--page-oracles', str(folder / 'pages.json'),
                '--buffer-proof', str(folder / 'proof.json')]
    result = subprocess.run(cmd, capture_output=True, text=True)
    (folder / 'command.json').write_text(json.dumps(cmd))
    (folder / 'output.log').write_text(result.stdout + result.stderr)
    assert (result.returncode == 0) == expected, (name, result.stdout, result.stderr)

run('legal-different-boundary-members')
run('strict-default-stays-strict', expected=False, amended=False)
for name, rows in [
    ('outside-boundary-group', [universe[1], universe[3], universe[6]]),
    ('invented-full-row-value', [['e', '8', 'changed'], universe[3], universe[6]]),
    ('wrong-group-cardinality', [universe[4], universe[5], universe[6]]),
    ('duplicate-over-multiplicity', [universe[4], universe[4], universe[6]]),
    ('wrong-order-between-groups', [universe[6], universe[3], universe[4]]),
    ('missing-page-row', [universe[4], universe[6]]),
]:
    changed = copy.deepcopy(b)
    changed.update(rows=rows, fingerprint=fingerprint(rows), returned=len(rows))
    run(name, target=changed, expected=False)

for name, field, value in [('target-work-increase', 'rows_processed', 31),
                         ('wrong-row-evidence-fingerprint', 'fingerprint', digest('wrong'))]:
    changed = copy.deepcopy(b)
    changed[field] = value
    run(name, target=changed, expected=False)
for key in snap:
    changed = copy.deepcopy(proof)
    changed['candidate'][key] = digest('changed')
    run('changed-' + key, identity=changed, expected=False)
for name, field, value in [('changed-unchanged-sql', 'sql_hash', digest('different')),
                         ('changed-estimated-plan', 'estimated_plan', [{'Plan': {'Plan Rows': 26}}])]:
    changed = copy.deepcopy(other_after)
    changed[field] = value
    run(name, unchanged=changed, expected=False)
changed = copy.deepcopy(proof)
changed['target_read_only'] = False
run('write-target-no-exemption', identity=changed, expected=False)
changed = copy.deepcopy(proof)
changed['variance']['0/other.Read/boundary'] = [other, other, other]
run('no-observed-buffer-variance', identity=changed, expected=False)

# Full groups inside a page cannot lose a member even when boundary members vary.
old_a, old_b, old_oracle = a, b, oracle
a = metric('target.Read', universe[1:7], 'baseline')
b = metric('target.Read', universe[1:7], 'candidate')
a['args'] = b['args'] = ['1', '6']
oracle = dict(old_oracle, args=['1', '6'])
run('complete-interior-group', target=b, page=oracle)
bad = copy.deepcopy(b)
bad['rows'][2] = universe[4]
bad['fingerprint'] = fingerprint(bad['rows'])
run('lost-interior-member', target=bad, page=oracle, expected=False)
bad = copy.deepcopy(b)
bad['rows'][0] = ['b', '9', 'null']
bad['fingerprint'] = fingerprint(bad['rows'])
run('null-is-not-string', target=bad, page=oracle, expected=False)

(out / 'summary.json').write_text(json.dumps({'status': 'passed', 'binary': binary,
    'checks': len(list(out.glob('*/gate.json'))),
    'boundary': 'Real CLI with synthetic measured-artifact inputs; live PostgreSQL collection checked separately.'}, indent=2))
print('PASS: amended qualification CLI and adversarial counterexamples')

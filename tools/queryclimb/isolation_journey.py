#!/usr/bin/env python3
import json
import pathlib
import subprocess
import sys
import uuid

binary, root, evidence, template = sys.argv[1:]
out = pathlib.Path(evidence)
out.mkdir(parents=True, exist_ok=True)
prefix = 'qc_attempt_identity_' + uuid.uuid4().hex[:12]

# Failure modes: the collector silently ignores changed data, schema, column
# statistics targets or planner settings; evidence is not collected from real DBs.
def call(name, *args, ok=True):
    p = subprocess.run([binary, '--root', root, *args], capture_output=True, text=True)
    (out / (name + '.log')).write_text(p.stdout + p.stderr)
    assert (p.returncode == 0) == ok, (name, p.stdout, p.stderr)
    return p

def sql(db, statement):
    p = subprocess.run(['docker', 'exec', '-i', 'tadoku-query-climb', 'psql', '-X', '-qAt',
                        '-U', 'tadoku_owner', '-d', db, '-v', 'ON_ERROR_STOP=1'],
                       input=statement, capture_output=True, text=True)
    assert p.returncode == 0, p.stderr

created = []
for scenario in ['same', 'data', 'schema', 'statistics', 'settings']:
    a, b = prefix + '_' + scenario + '_a', prefix + '_' + scenario + '_b'
    for db in [a, b]:
        call('clone-' + db, 'clone', template, db)
        created.append(db)
    if scenario == 'data':
        sql(b, "update languages set name = name || ' changed' where code = 'jpn'")
    elif scenario == 'schema':
        sql(b, 'alter table logs add column identity_probe integer')
    elif scenario == 'statistics':
        sql(b, 'alter table logs alter column created_at set statistics 211')
    elif scenario == 'settings':
        sql(b, 'alter database ' + b + " set work_mem = '3MB'")
    call(scenario, 'isolation', '--baseline-name', a, '--candidate-name', b,
         '--candidate-root', root, '--target', 'logs.ListLogsForContest',
         '--output', str(out / (scenario + '.json')), ok=scenario == 'same')

(out / 'summary.json').write_text(json.dumps({'status': 'passed', 'binary': binary,
    'template': template, 'databases_retained': created,
    'checks': ['same clones', 'changed data', 'changed schema', 'changed statistics', 'changed settings']}, indent=2))
print('PASS: live PostgreSQL identity collector and counterexamples')

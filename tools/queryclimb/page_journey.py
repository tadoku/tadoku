#!/usr/bin/env python3
import copy
import hashlib
import json
import pathlib
import subprocess
import sys
import uuid

binary, root, evidence, template, case_file = sys.argv[1:]
out = pathlib.Path(evidence)
out.mkdir(parents=True, exist_ok=True)
database = 'qc_attempt_page_' + uuid.uuid4().hex[:12]
case = next(c for c in json.loads(pathlib.Path(case_file).read_text())
            if c['query'] == 'logs.ListLogsForContest' and c['name'] == 'large-deep-all')
case['include_rows'] = True
(out / 'case.json').write_text(json.dumps([case]))

# Failure modes: wire row evidence does not match its fingerprint, the full-row
# oracle changes NULLs or counts, or the gate accepts a row absent from the DB.
def call(name, *args, ok=True):
    command = [binary, '--root', root, *args]
    p = subprocess.run(command, capture_output=True, text=True)
    (out / (name + '.command.json')).write_text(json.dumps(command))
    (out / (name + '.log')).write_text(p.stdout + p.stderr)
    assert (p.returncode == 0) == ok, (name, p.stdout, p.stderr)
    return p.stdout

def encoded(value):
    return json.dumps(value, separators=(',', ':'), ensure_ascii=False)

def fingerprint(rows):
    groups = []
    for row in rows:
        key = [row[col] for col in case['tie_columns']]
        if not groups or groups[-1][0] != key:
            groups.append((key, []))
        groups[-1][1].append(encoded(row))
    value = encoded([row for _, group in groups for row in sorted(group)])
    return hashlib.sha256(value.encode()).hexdigest()

call('clone', 'clone', template, database)
raw = call('inspect', 'inspect', '--name', database, '--cases', str(out / 'case.json'))
baseline = json.loads(raw)
(out / 'baseline.json').write_text(raw)
call('oracle', 'page-evidence', '--name', database, '--cases', str(out / 'case.json'),
     '--offset-arg', '3', '--limit-arg', '4', '--output', str(out / 'oracle.json'))
oracle = json.loads((out / 'oracle.json').read_text())[0]
assert oracle['sql_hash'] == baseline[0]['sql_hash']
assert fingerprint(baseline[0]['rows']) == baseline[0]['fingerprint']
candidate = copy.deepcopy(baseline)
used = {encoded(row) for row in candidate[0]['rows']}
found = False
for i, selected in enumerate(candidate[0]['rows']):
    key = [selected[col] for col in case['tie_columns']]
    alternative = next((row for row in oracle['rows']
                        if [row[col] for col in case['tie_columns']] == key
                        and encoded(row) not in used), None)
    if alternative is not None:
        candidate[0]['rows'][i] = alternative
        found = True
        break
assert found, 'fixture must cut an ORDER BY tie group'
candidate[0]['fingerprint'] = fingerprint(candidate[0]['rows'])
(out / 'candidate.json').write_text(json.dumps(candidate))
args = ['gate', '--mode', 'guard', '--target', case['query'],
        '--baseline', str(out / 'baseline.json'), '--candidate', str(out / 'candidate.json')]
call('strict-rejects', *args, ok=False)
call('legal-alternative', *args, '--page-oracles', str(out / 'oracle.json'))
candidate[0]['rows'][0][0] = '00000000-0000-0000-0000-000000000000'
candidate[0]['fingerprint'] = fingerprint(candidate[0]['rows'])
(out / 'candidate.json').write_text(json.dumps(candidate))
call('invented-row-rejected', *args, '--page-oracles', str(out / 'oracle.json'), ok=False)
(out / 'summary.json').write_text(json.dumps({'status': 'passed', 'database_retained': database,
    'template': template, 'query': case['query'], 'case': case['name'],
    'boundary': 'Real PostgreSQL rows/oracle; alternate and invalid candidate measurement artifacts.'}, indent=2))
print('PASS: real PostgreSQL page evidence, legal tie alternative and invalid-row rejection')

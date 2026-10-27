#!/usr/bin/env python3
import json
import pathlib
import subprocess
import sys
import fcntl
import time

binary, root, evidence, database = sys.argv[1:]
out = pathlib.Path(evidence)
out.mkdir(parents=True, exist_ok=True)

def run(*args, ok=True, contains=None):
    p = subprocess.run([binary, '--root', root, *args], text=True, capture_output=True)
    (out / ('-'.join(args[:1]) + '.log')).write_text(p.stdout + p.stderr)
    assert (p.returncode == 0) == ok, (args, p.returncode, p.stdout, p.stderr)
    if contains is not None:
        assert contains in p.stderr, p.stderr
    return p.stdout

# Failure modes: missing nested constants, stale source/generated SQL, unsafe DSNs,
# tenant leakage, execution 6 replanning, changed rows, ignored ordering, leaked writes.
with open('/tmp/tadoku-query-climb.lock', 'a') as lock:
    fcntl.flock(lock, fcntl.LOCK_EX)
    queued = subprocess.Popen([binary, '--root', root, 'inventory'], text=True, stdout=subprocess.PIPE)
    time.sleep(0.1)
    assert queued.poll() is None, 'operation bypassed the measurement lock'
    fcntl.flock(lock, fcntl.LOCK_UN)
    inventory_output, _ = queued.communicate(timeout=30)
    assert queued.returncode == 0
(out / 'inventory.log').write_text(inventory_output)
queries = json.loads(inventory_output)
assert len(queries) == 157
assert len({q['key'] for q in queries}) == 157
assert any(q['key'] == 'identityreads.FindAccount' for q in queries)
assert all(isinstance(q['tables'], list) and q['fingerprint'] for q in queries)
run('inspect', '--database', 'postgres://localhost/not-local', ok=False,
    contains='only the owned query-climb loopback instance')
run('selfcheck', '--name', database)
print('PASS: 157-query inventory and real PostgreSQL harness journey')

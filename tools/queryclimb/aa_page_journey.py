#!/usr/bin/env python3
import argparse
import copy
import hashlib
import json
import os
import pathlib
import selectors
import signal
import subprocess
import time

parser = argparse.ArgumentParser(description='Live legal-page binding preflight; no completed timing claim.')
for name in ['binary', 'root', 'evidence', 'source', 'candidate', 'cases', 'oracle']:
    parser.add_argument(name)
parser.add_argument('--expect-old-binding-error', action='store_true')
args = parser.parse_args()
out = pathlib.Path(args.evidence)
out.mkdir(parents=True, exist_ok=False)
oracle_path = pathlib.Path(args.oracle)
original_bytes = oracle_path.read_bytes()
oracles = json.loads(original_bytes)
cases = json.loads(pathlib.Path(args.cases).read_text())
assert len(cases) == len(oracles) == 1
assert cases[0]['include_rows'] and oracles[0]['database'] == args.source
marker = f"case {cases[0]['query']}/{cases[0]['name']}: "
binding_error = 'page oracle is not bound to the baseline query and exact case'
checks = []


def run(name, pages, candidate, accepted):
    path = out / (name + '-oracle.json')
    path.write_text(json.dumps(pages))
    command = [args.binary, '--root', args.root, 'baseline', '--name', args.source,
               '--candidate-name', candidate, '--candidate-root', args.root,
               '--cases', args.cases, '--page-oracles', str(path)]
    record = dict(name=name, command=command, preflight_only=True,
                  completed_timing_pairs=False, expected_marker=accepted)
    (out / (name + '-command.json')).write_text(json.dumps(record, indent=2))
    start = time.monotonic()
    reached = False
    cancelled = False
    stderr = ''
    with (out / (name + '.stdout')).open('ab') as stdout, (out / (name + '.stderr')).open('ab') as log:
        process = subprocess.Popen(command, stdout=stdout, stderr=subprocess.PIPE)
        record['pid'] = process.pid

        def cancel():
            proc = pathlib.Path('/proc') / str(process.pid)
            identity = dict(pid=process.pid, uid=proc.stat().st_uid,
                            executable=os.readlink(proc / 'exe'),
                            command=(proc / 'cmdline').read_bytes().split(b'\0')[:-1],
                            status=(proc / 'status').read_text())
            assert identity['uid'] == os.getuid()
            assert f'PPid:\t{os.getpid()}\n' in identity['status']
            assert identity['executable'] == str(pathlib.Path(args.binary).resolve())
            assert identity['command'] == [os.fsencode(arg) for arg in command]
            record['cancellation_identity'] = dict(identity, command=command)
            process.send_signal(signal.SIGTERM)

        selector = selectors.DefaultSelector()
        selector.register(process.stderr, selectors.EVENT_READ)
        try:
            while selector.get_map():
                if time.monotonic() - start > 180:
                    raise TimeoutError(name)
                for key, _ in selector.select(timeout=1):
                    chunk = os.read(key.fileobj.fileno(), 65536)
                    if not chunk:
                        selector.unregister(key.fileobj)
                        continue
                    log.write(chunk)
                    log.flush()
                    stderr += chunk.decode()
                    if marker in stderr and not cancelled:
                        reached = True
                        cancel()
                        cancelled = True
            record.update(exit=process.wait(timeout=10), reached_marker=reached,
                          deliberately_cancelled=cancelled, elapsed_seconds=time.monotonic() - start)
        finally:
            selector.close()
            process.stderr.close()
            if process.poll() is None:
                cancel()
                process.wait(timeout=10)
    checks.append(record)
    (out / 'checks.json').write_text(json.dumps(checks, indent=2))
    if accepted:
        assert reached and cancelled and record['exit'] == -signal.SIGTERM, record
    else:
        assert not reached and record['exit'] == 1 and binding_error in stderr, record
    assert oracle_path.read_bytes() == original_bytes


# Failure modes: the harness clone loses a valid source binding; unrelated DBs
# or SQL are rebound; explicit A/B loses its exact baseline binding.
run('source-aa', oracles, args.source, not args.expect_old_binding_error)
if not args.expect_old_binding_error:
    unrelated = copy.deepcopy(oracles)
    unrelated[0]['database'] = args.candidate
    run('unrelated-aa-rejected', unrelated, args.source, False)
    wrong_sql = copy.deepcopy(oracles)
    wrong_sql[0]['sql_hash'] = '0' * 64
    run('wrong-sql-aa-rejected', wrong_sql, args.source, False)
    run('source-ab', oracles, args.candidate, True)
    run('unrelated-ab-rejected', unrelated, args.candidate, False)
    run('wrong-sql-ab-rejected', wrong_sql, args.candidate, False)

(out / 'summary.json').write_text(json.dumps(dict(
    status='passed', binary=args.binary,
    binary_sha256=hashlib.sha256(pathlib.Path(args.binary).read_bytes()).hexdigest(),
    oracle_sha256=hashlib.sha256(original_bytes).hexdigest(), checks=len(checks),
    old_binding_failure=args.expect_old_binding_error, preflight_only=True,
    boundary='Success stops the owned CLI at the post-validator marker; no 15-pair timing claim. '
             'SIGTERM retains temporary clones. Source oracle bytes remain unchanged.'), indent=2))
print('PASS: live legal-page binding preflight (timing deliberately cancelled)')

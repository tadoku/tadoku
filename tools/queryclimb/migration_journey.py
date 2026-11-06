#!/usr/bin/env python3
import json
import pathlib
import subprocess
import sys
import uuid

binary, root, evidence, template = sys.argv[1:]
out = pathlib.Path(evidence)
out.mkdir(parents=True, exist_ok=True)
root = str(pathlib.Path(root).resolve())
run_id = uuid.uuid4().hex[:12]
migrations_dir = 'services/tadoku-api/migrations'

# Failure modes: attempt re-applies or skips migrations by a fixed number instead
# of comparing candidate and baseline; it accepts a baseline that is dirty or not
# origin/main, a template built from other migrations or without a manifest, or a
# candidate that edits, removes, misnumbers or adds anything but one standalone
# concurrent index; or it applies a new index to the baseline clone.
def git(*args, cwd=root):
    p = subprocess.run(['git', *args], cwd=cwd, capture_output=True, text=True)
    assert p.returncode == 0, (args, p.stderr)
    return p.stdout.strip()

def sql(db, statement):
    p = subprocess.run(['docker', 'exec', '-i', 'tadoku-query-climb', 'psql', '-X', '-qAt',
                        '-U', 'tadoku_owner', '-d', db, '-v', 'ON_ERROR_STOP=1'],
                       input=statement, capture_output=True, text=True)
    assert p.returncode == 0, p.stderr

worktrees = []
def checkout(name, revision='HEAD'):
    path = out / ('checkout-' + name)
    git('worktree', 'add', '--detach', str(path), git('rev-parse', revision))
    worktrees.append(path)
    return path

def migration_files(path):
    return sorted(p.name for p in (pathlib.Path(path) / migrations_dir).glob('*.sql'))

numbers = [int(name.split('_', 1)[0]) for name in migration_files(root)]
next_number = max(numbers) + 1
index_name = 'qc_journey_jobs_unfinished_' + run_id
concurrent_index = (
    f'create index concurrently {index_name}\n'
    '  on jobs (task_type)\n'
    '  include (tenant, state, next_attempt_at, lease_expires_at)\n'
    "  where state <> 'completed';\n"
)

def add_migration(path, number, body, name='journey_probe'):
    file = pathlib.Path(path) / migrations_dir / f'{number:04d}_{name}.up.sql'
    file.write_text(body)
    file.with_name(file.name.replace('.up.sql', '.down.sql')).write_text(
        f'drop index concurrently if exists {index_name};\n')
    return file.name

cases = out / 'cases.json'
cases.write_text(json.dumps([{
    'name': 'mixed-states', 'query': 'jobqueue.Stats', 'tenant': 'tadoku/prod',
    'args': ['climb.work.v1', 'worker'], 'tie_columns': [],
}], indent=1))
aa = out / 'unused-aa.json'
aa.write_text('[]')

def attempt(name, baseline, candidate, templates=None, rejects=None):
    templates = templates or [template, template]
    p = subprocess.run([binary, '--root', str(baseline), 'attempt',
                        '--candidate-root', str(candidate),
                        '--templates', ','.join(templates),
                        '--cases', f'{cases},{cases}',
                        '--target', 'jobqueue.Stats',
                        '--output', str(out / ('attempt-' + name)),
                        '--aa-calibration', f'{aa},{aa}'],
                       capture_output=True, text=True)
    (out / (name + '.log')).write_text(p.stdout + p.stderr)
    if rejects:
        assert p.returncode != 0 and rejects in p.stderr, (name, rejects, p.stderr[-2000:])
    return p

def used_indexes(path):
    return {index for sample in json.loads(path.read_text()) for index in sample['indexes']}

checks = []
try:
    # Current main already holds migrations that are not index builds. A new
    # index must reach the candidate clone only, and the attempt must get past
    # migration planning to case inspection.
    candidate = checkout('candidate-index')
    added = add_migration(candidate, next_number, concurrent_index)
    p = attempt('candidate-index', root, candidate)
    plan_path = out / 'attempt-candidate-index' / 'migrations.json'
    assert plan_path.exists(), ('no migration plan', p.stderr[-2000:])
    plan = json.loads(plan_path.read_text())
    assert plan['candidate_only'] == [added], plan
    baseline_indexes = used_indexes(out / 'attempt-candidate-index' / 'baseline-0.json')
    candidate_indexes = used_indexes(out / 'attempt-candidate-index' / 'candidate-0.json')
    assert index_name in candidate_indexes, candidate_indexes
    assert index_name not in baseline_indexes, baseline_indexes
    checks.append('new concurrent index applied to the candidate clone only')

    changed = checkout('candidate-changed')
    first = migration_files(changed)[0]
    path = pathlib.Path(changed) / migrations_dir / first
    path.write_text(path.read_text() + '\n-- changed\n')
    attempt('candidate-changed', root, changed, rejects='candidate changed existing migration')
    checks.append('edited existing migration rejected')

    removed = checkout('candidate-removed')
    (pathlib.Path(removed) / migrations_dir / migration_files(removed)[-1]).unlink()
    attempt('candidate-removed', root, removed, rejects='candidate removed migration')
    checks.append('removed migration rejected')

    misnumbered = checkout('candidate-misnumbered')
    add_migration(misnumbered, max(numbers), concurrent_index, name='journey_reused_number')
    attempt('candidate-misnumbered', root, misnumbered,
            rejects='new migration must be numbered after')
    checks.append('new migration numbered inside the baseline range rejected')

    for name, body in [
        ('candidate-not-index', 'alter table jobs add column journey_probe integer;\n'),
        ('candidate-blocking-index', f'create index {index_name} on jobs (task_type);\n'),
        ('candidate-two-statements', concurrent_index + 'select 1;\n'),
    ]:
        other = checkout(name)
        add_migration(other, next_number, body)
        attempt(name, root, other, rejects='new migration must be one create index concurrently statement')
        checks.append(name + ' rejected')

    dirty = checkout('baseline-dirty')
    add_migration(dirty, next_number, concurrent_index)
    attempt('baseline-dirty', dirty, candidate, rejects='baseline must match origin/main')
    checks.append('dirty baseline rejected')

    last_change = git('rev-list', '-1', 'HEAD', '--', migrations_dir)
    old = checkout('baseline-old', last_change + '^')
    attempt('baseline-old', old, candidate, rejects='baseline must match origin/main')
    checks.append('baseline behind origin/main rejected')

    stale, bare = f'qc_journey_stale_{run_id}', f'qc_journey_bare_{run_id}'
    for db in [stale, bare]:
        p = subprocess.run([binary, '--root', root, 'clone', template, db], capture_output=True, text=True)
        assert p.returncode == 0, p.stderr
    sql(stale, 'delete from queryclimb.template_migrations '
               'where name = (select max(name) from queryclimb.template_migrations)')
    sql(bare, 'drop schema queryclimb cascade')
    attempt('template-stale', root, candidate, templates=[stale, stale],
            rejects='template migrations differ from baseline')
    attempt('template-bare', root, candidate, templates=[bare, bare],
            rejects='template has no migration manifest')
    checks.append('stale and manifest-less templates rejected')
finally:
    for path in worktrees:
        subprocess.run(['git', 'worktree', 'remove', '--force', str(path)], cwd=root, capture_output=True)

(out / 'summary.json').write_text(json.dumps({
    'status': 'passed', 'binary': binary, 'root': root, 'root_revision': git('rev-parse', 'HEAD'),
    'template': template, 'journey_databases_retained': [stale, bare], 'checks': checks,
}, indent=2))
print('PASS: attempt migration planning against origin/main and a live template')

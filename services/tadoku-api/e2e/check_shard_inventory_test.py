import pathlib, subprocess, tempfile
checker = pathlib.Path(__file__).with_name('check_shard_inventory.py')
with tempfile.TemporaryDirectory() as d:
    root = pathlib.Path(d)
    def xml(name, cases):
        p = root / name
        p.write_text('<testsuites><testsuite>' + ''.join(f'<testcase classname="e2e" name="{c}"/>' for c in cases) + '</testsuite></testsuites>')
        return str(p)
    baseline = xml('baseline.xml', ['TestA', 'TestA/child', 'TestB'])
    good = [xml('s0.xml', ['TestA', 'TestA/child']), xml('s1.xml', ['TestB'])]
    bad = [xml('missing.xml', ['TestA', 'TestB']), xml('duplicate.xml', ['TestA', 'TestA/child', 'TestA/child']), xml('renamed.xml', ['TestA', 'TestA/child', 'TestC'])]
    for paths, want in [(good, 0)] + [([p], 1) for p in bad]:
        result = subprocess.run(['python3', str(checker), '--baseline', baseline, '--shards', *paths], capture_output=True, text=True)
        assert result.returncode == want, (paths, result.returncode, result.stdout, result.stderr)
    for tag in ['skipped', 'failure', 'error']:
        p = root / (tag + '.xml')
        p.write_text(f'<testsuite><testcase classname="e2e" name="TestA"><{tag}/></testcase></testsuite>')
        result = subprocess.run(['python3', str(checker), '--baseline', baseline, '--shards', str(p)], capture_output=True, text=True)
        assert result.returncode == 1, (tag, result)
print('inventory positive, missing, duplicate, renamed, skipped, failed and error controls passed')

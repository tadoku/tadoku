#!/usr/bin/env python3
"""Real Bazelisk/Bazel integration."""

import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import signal
import subprocess
import sys
import time
import uuid


FAILURE_CASES = [
    "build credentials",
    "worktrees/TMPDIRs bypass the lock",
    "servers/targets retain the lock",
    "resource defaults and overrides",
    "run argv, stdio, exit, cwd, runfiles, environment",
    "nested Bazel bypass/deadlock",
    "CLI/RC/config norun/script_path executes target",
    "build/target signals and failure cleanup",
    "test URLs or Docker config lost",
]

sys.dont_write_bytecode = True
from bazel_wrapper_fixture import ACTION, DEFS, PROBE


def wait_for(predicate, description, timeout=90):
    deadline = time.monotonic() + timeout
    while not predicate():
        if time.monotonic() > deadline:
            raise AssertionError("timeout: " + description)
        time.sleep(.1)


def effective_options(path):
    for line in path.read_text().splitlines():
        command = json.loads(line).get("structuredCommandLine", {})
        if command.get("commandLineLabel") == "canonical":
            return {option["optionName"]: option.get("optionValue", "")
                    for section in command["sections"]
                    for option in section.get("optionList", {}).get("option", [])}
    raise AssertionError("missing canonical options")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--evidence-dir", required=True, type=Path)
    parser.add_argument("--unwrapped", action="store_true")
    parser.add_argument("--prepare-only", action="store_true")
    options = parser.parse_args()
    repo = Path(__file__).resolve().parents[2]
    evidence = options.evidence_dir.resolve()
    evidence.mkdir(parents=True, exist_ok=True)
    bazel = shutil.which("bazel")
    assert bazel, "bazel (Bazelisk) required"
    version = (repo / ".bazelversion").read_text()
    hook = repo / "tools/bazel"
    home = evidence / "home"
    (home / ".docker").mkdir(parents=True, exist_ok=True)
    (home / ".docker/config.json").write_text('{"auths":{}}\n')
    tmp = evidence / "tmp"
    tmp.mkdir(exist_ok=True)
    env = {key: os.environ[key] for key in ("PATH", "USER", "LANG", "LC_ALL") if key in os.environ}
    env.update(HOME=str(home), TMPDIR=str(tmp), TERM="dumb",
               BAZELISK_HOME=os.environ.get("BAZELISK_HOME", str(Path.home() / ".cache/bazelisk")),
               GH_TOKEN="wrapper-canary-" + uuid.uuid4().hex, NEW_PROVIDER_SECRET="unknown-canary-" + uuid.uuid4().hex,
               TADOKU_TEST_POSTGRES_URL="postgres://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable",
               TADOKU_TEST_PGBOUNCER_URL="postgres://postgres:postgres@127.0.0.1:6432/postgres?sslmode=disable",
               TADOKU_TEST_VALKEY_URL="redis://127.0.0.1:6379", TADOKU_TEST_EMAIL_GATEWAY_RENDER="fixture")
    bases = [evidence / "output-a", evidence / "output-b"]
    workspaces = [evidence / "workspace-a", evidence / "workspace-b"]
    seed = evidence / "fixture-repo"
    if not seed.exists():
        seed.mkdir()
        subprocess.run(["git", "init", "-q", str(seed)], check=True)
        (seed / "MODULE.bazel").write_text('module(name = "wrapper_fixture")\n')
        subprocess.run(["git", "-C", str(seed), "add", "MODULE.bazel"], check=True)
        subprocess.run(["git", "-C", str(seed), "-c", "user.name=Wrapper check", "-c",
                        "user.email=wrapper@example.invalid", "commit", "-qm", "Fixture"], check=True)
        for workspace in workspaces:
            subprocess.run(["git", "-C", str(seed), "worktree", "add", "--detach", str(workspace)],
                           check=True, capture_output=True)
    for workspace in workspaces:
        workspace.mkdir(exist_ok=True)
        (workspace / "MODULE.bazel").write_text('module(name = "wrapper_fixture")\n')
        (workspace / ".bazelversion").write_text(version)
        (workspace / "defs.bzl").write_text(DEFS)
        (workspace / "probe.py").write_text(PROBE)
        (workspace / "action.py").write_text(ACTION)
        (workspace / "payload.txt").write_text("runfiles preserved")
        (workspace / "probe.py").chmod(0o755)
        (workspace / "action.py").chmod(0o755)
        (workspace / "release").touch()
        (workspace / "BUILD.bazel").write_text(
            'load(":defs.bzl", "probe", "action")\n'
            'probe(name="probe", src="probe.py", data=["payload.txt", ":action"])\n'
            'action(name="action", src="action.py", directory=' + json.dumps(str(workspace)) + ')\n')
        if not options.unwrapped:
            (workspace / "tools").mkdir(exist_ok=True)
            if not (workspace / "tools/bazel").is_symlink():
                (workspace / "tools/bazel").symlink_to(hook)
    manifest = {"revision": subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=repo, text=True).strip(),
                "command": sys.argv, "failure_cases": FAILURE_CASES, "bazel": bazel,
                "cli_version": None if options.prepare_only else subprocess.check_output(
                    [bazel, "--version"], cwd=workspaces[0], env=env, text=True).strip(),
                "bazelisk_sha256": hashlib.sha256(Path(bazel).read_bytes()).hexdigest(),
                "version": version.strip(), "unwrapped": options.unwrapped,
                "check_sha256": hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
                "hook_sha256": hashlib.sha256(hook.read_bytes()).hexdigest() if hook.exists() else None}
    if not options.prepare_only:
        assert manifest["cli_version"] == "bazel " + version.strip()
    for source in (Path(__file__), Path(__file__).with_name("bazel_wrapper_fixture.py"), hook):
        if source.exists():
            shutil.copyfile(source, evidence / (source.name + ".source"))
    (evidence / "setup.json").write_text(json.dumps(manifest, indent=2) + "\n")
    (evidence / "tested.diff").write_bytes(subprocess.check_output(["git", "diff"], cwd=repo))
    if options.prepare_only:
        print("Prepared real Bazel fixtures:", evidence)
        return
    processes = []
    counter = 0
    checks = []
    startup = ["--host_jvm_args=-Xmx256m", "--host_jvm_args=-XX:ActiveProcessorCount=2", "--max_idle_secs=15"]
    def command(args, index=0, defaults=False):
        return [bazel, "--output_base", str(bases[index]), "--nosystem_rc", "--nohome_rc",
                *(startup if not defaults else []), *args]
    def run(args, index=0, expected=0, extra_env=None, stdin="", defaults=False):
        nonlocal counter
        counter += 1
        argv = command(args, index, defaults)
        result = subprocess.run(argv, cwd=workspaces[index], env=env | (extra_env or {}),
                                input=stdin, capture_output=True, text=True, timeout=180)
        (evidence / f"{counter:02}.log").write_text(result.stdout + result.stderr)
        with (evidence / "commands.jsonl").open("a") as log:
            log.write(json.dumps({"argv": argv, "cwd": str(workspaces[index]), "exit": result.returncode}) + "\n")
        assert result.returncode in (expected if isinstance(expected, tuple) else (expected,)), f"command {counter} returned {result.returncode}, expected {expected}; see evidence"
        return result
    def start(args, index=0, extra_env=None):
        nonlocal counter
        counter += 1
        log = (evidence / f"{counter:02}.log").open("w")
        child = subprocess.Popen(command(args, index), cwd=workspaces[index], env=env | (extra_env or {}),
                                 stdout=log, stderr=log, stdin=subprocess.DEVNULL, start_new_session=True)
        log.close()
        processes.append(child)
        return child
    def clean_logs():
        for path in [*evidence.glob("*.log"), *(p for base in bases for p in base.glob("java.log*"))]:
            contents = path.read_bytes()
            for name in ("GH_TOKEN", "NEW_PROVIDER_SECRET"):
                assert env[name].encode() not in contents, f"credential {name} reached {path.name}"
    def pass_check(name):
        checks.append(name)
        print("PASS:", name, flush=True)
    try:
        out = evidence / "runtime.json"
        preflight = evidence / "preflight.jsonl"
        run(["build", "--incompatible_autoload_externally=", "//:probe"])
        run(["run", "--incompatible_autoload_externally=", "--noanalyze", "--nobuild",
             "--build_event_json_file=" + str(preflight), "//:probe", "--", "record", str(out)], expected=(0, 2))
        assert not out.exists(), "noanalyze executed cached target"
        assert effective_options(preflight)["analyze"] in ("0", "false")
        pass_check("cached noanalyze")
        bep = evidence / "build.jsonl"
        run(["build", "--incompatible_autoload_externally=", "--action_env=GH_TOKEN",
             "--action_env=NEW_PROVIDER_SECRET", "--action_env=TADOKU_TEST_POSTGRES_URL",
             "--build_event_json_file=" + str(bep), "//:action"])
        clean_logs()
        action = json.loads((workspaces[0] / "bazel-bin/action.json").read_text())
        assert action["GH_TOKEN"] is None and action["NEW_PROVIDER_SECRET"] is None, "credentials reached build action"
        assert action["TADOKU_TEST_POSTGRES_URL"] == env["TADOKU_TEST_POSTGRES_URL"]
        pass_check("build scrub/test URL")
        assert effective_options(bep)["local_resources"] == "memory=HOST_RAM*0.25"
        run(["build", "--incompatible_autoload_externally=", "--local_resources=memory=128",
             "--build_event_json_file=" + str(bep), "//:action"])
        assert effective_options(bep)["local_resources"] == "memory=128"
        run(["info", "server_pid"])
        assert "-Xmx256m" in (bases[0] / "server/cmdline").read_text()
        pass_check("resource overrides")
        args = ["--", "", "space and 'quotes'", '"double"', "line\nbreak", "--script_path=not-an-option"]
        result = run(["run", "--incompatible_autoload_externally=", "--run_env=GH_TOKEN",
                     "--run_env==NEW_PROVIDER_SECRET", "//:probe", "--", "exit", str(out), *args],
                     stdin="stdin retained\n", expected=17)
        observed = json.loads(out.read_text())
        assert observed["argv"] == args and observed["stdin"] == "stdin retained\n"
        assert observed["runfile"] == "runfiles preserved" and observed["cwd"].endswith(".runfiles/_main")
        assert observed["env"]["BUILD_WORKSPACE_DIRECTORY"] == str(workspaces[0])
        assert observed["env"]["BUILD_WORKING_DIRECTORY"] == str(workspaces[0])
        assert observed["env"]["GH_TOKEN"] == env["GH_TOKEN"] and observed["docker"]
        assert observed["env"]["NEW_PROVIDER_SECRET"] is None
        assert not observed["env"]["BAZEL_REAL"] and not observed["env"]["BAZELISK_SKIP_WRAPPER"]
        assert "probe stdout" in result.stdout and "probe stderr" in result.stderr
        assert not list(tmp.glob("tadoku-bazel-*")), "private scripts survived target exec"
        pass_check("run argv/stdio/exit/cwd/runfiles/env/Docker")
        for flags in (["--norun"], ["--run=false"], ["--config=skip"]):
            (workspaces[0] / ".bazelrc").write_text('import %workspace%/imported.rc\n')
            (workspaces[0] / "imported.rc").write_text('run:skip --norun\n')
            out.unlink(missing_ok=True)
            run(["run", "--incompatible_autoload_externally=", *flags, "//:probe", "--", "record", str(out)])
            assert not out.exists(), "norun executed target"
        (workspaces[0] / "imported.rc").write_text('run --norun\n')
        run(["run", "--incompatible_autoload_externally=", "//:probe", "--", "record", str(out)])
        assert not out.exists(), "rc norun executed target"
        run(["run", "--incompatible_autoload_externally=", "--run", "//:probe", "--", "record", str(out)])
        assert out.exists(), "CLI run did not override rc norun"
        (workspaces[0] / "imported.rc").write_text("")
        for via_rc in (False, True):
            script = evidence / ("caller-rc.sh" if via_rc else "caller-cli.sh")
            flags = [] if via_rc else ["--script_path", str(script)]
            (workspaces[0] / "imported.rc").write_text("run --script_path=" + str(script) + "\n" if via_rc else "")
            out.unlink()
            run(["run", "--incompatible_autoload_externally=", *flags, "//:probe", "--", "record", str(out)])
            assert script.exists() and not out.exists(), "script_path executed target or lost script"
            subprocess.run([str(script)], env=env, input="", text=True, check=True, capture_output=True)
            assert out.exists()
        (workspaces[0] / ".bazelrc").write_text("")
        pass_check("CLI/RC/config run controls")
        user_bep = evidence / "caller-events.jsonl"
        run(["run", "--incompatible_autoload_externally=", "--build_event_json_file=" + str(user_bep),
             "//:probe", "--", "record", str(out)], extra_env={"DOCKER_CONFIG": str(home / ".docker")})
        assert effective_options(user_bep).get("analyze") != "0"
        assert json.loads(out.read_text())["docker"]
        pass_check("BEP/Docker override")
        run(["run", "--incompatible_autoload_externally=", "//:missing"], expected=1)
        assert not list(tmp.glob("tadoku-bazel-*")), "failed build leaked private scripts"
        pass_check("failure cleanup/exit")
        lock = Path("/tmp") / f"tadoku-bazel-{os.getuid()}.lock"
        held = evidence / "held"
        holder = subprocess.Popen(["flock", "-o", str(lock), sys.executable, "-c",
            "import pathlib,sys,time; pathlib.Path(sys.argv[1]).touch(); time.sleep(20)", str(held)],
            start_new_session=True)
        processes.append(holder)
        wait_for(held.exists, "held production lock")
        queued = start(["info", "server_pid"])
        wait_for(lambda: list(tmp.glob("tadoku-bazel-*")), "queued wrapper")
        os.killpg(queued.pid, signal.SIGINT)
        assert queued.wait(timeout=5) == 130
        assert not list(tmp.glob("tadoku-bazel-*")), "queued cancellation leaked files"
        os.killpg(holder.pid, signal.SIGTERM)
        holder.wait(timeout=5)
        pass_check("queued lock cancellation")

        ready = evidence / "target.pid"
        child = start(["run", "--incompatible_autoload_externally=", "//:probe", "--", "wait", str(ready)])
        wait_for(ready.exists, "long running target")
        lock = Path("/tmp") / f"tadoku-bazel-{os.getuid()}.lock"
        assert subprocess.run(["flock", "-n", str(lock), "true"]).returncode == 0, "target owns build lock"
        assert not list(tmp.glob("tadoku-bazel-*")), "private script survives long running target"
        assert child.pid == int(ready.read_text()), "wrapper did not exec target"
        os.killpg(child.pid, signal.SIGTERM)
        assert child.wait(timeout=10) == 42
        pass_check("target unlock/cleanup/signals")
        workspace = workspaces[0]
        (workspace / "release").unlink()
        (workspace / "started").unlink()
        (workspace / "action.py").write_text(ACTION + "\n# interrupt\n")
        child = start(["run", "--incompatible_autoload_externally=", "//:probe", "--", "record", str(out)])
        wait_for((workspace / "started").exists, "interruptible build")
        os.killpg(child.pid, signal.SIGTERM)
        assert child.wait(timeout=20) == 143
        assert not list(tmp.glob("tadoku-bazel-*")), "interrupted build leaked scripts"
        (workspace / "release").touch()
        pass_check("build SIGTERM cleanup")
        nested = evidence / "nested.json"
        run(["run", "--incompatible_autoload_externally=", "//:probe", "--", "nested", str(nested), str(bases[0])])
        assert json.loads(nested.read_text())["code"] == 0
        assert effective_options(Path(str(nested) + ".bep"))["local_resources"] == "memory=HOST_RAM*0.25"
        clean_logs()
        pass_check("nested hook")
        run(["build", "--incompatible_autoload_externally=", "//:action"], index=1)
        for workspace in workspaces:
            (workspace / "release").unlink()
            (workspace / "started").unlink(missing_ok=True)
            (workspace / "finished").unlink(missing_ok=True)
            (workspace / "action.py").write_text(ACTION + "\n")
        first = start(["build", "--incompatible_autoload_externally=", "//:action"])
        wait_for((workspaces[0] / "started").exists, "first action")
        other_tmp = evidence / "different-tmp"
        other_tmp.mkdir()
        second = start(["build", "--incompatible_autoload_externally=", "//:action"], 1,
                       {"TMPDIR": str(other_tmp)})
        time.sleep(1)
        assert not (workspaces[1] / "started").exists(), "two workspaces built simultaneously"
        (workspaces[0] / "release").touch()
        assert first.wait(timeout=90) == 0
        wait_for((workspaces[1] / "started").exists, "second action after unlock")
        (workspaces[1] / "release").touch()
        assert second.wait(timeout=90) == 0
        intervals = [{name: float((workspace / name).read_text()) for name in ("started", "finished")}
                     for workspace in workspaces]
        (evidence / "concurrency.json").write_text(json.dumps(intervals, indent=2) + "\n")
        assert intervals[0]["finished"] <= intervals[1]["started"], "worktree actions overlapped"
        assert subprocess.run(["fuser", str(lock)], capture_output=True).returncode == 1, "server inherited lock descriptor"
        pass_check("worktree/TMPDIR lock/no leaked FD")
        run(["shutdown"])
        result = run(["info"], defaults=True)
        assert "max-heap-size:" in result.stdout
        server_args = (bases[0] / "server/cmdline").read_text()
        assert "-Xmx1024m" in server_args
        assert "--max_idle_secs=60" in server_args
        run(["shutdown"], defaults=True)
        result = run(["--max_idle_secs=2", "info", "server_pid"])
        idle_pid = int(result.stdout.strip())
        def stopped():
            try:
                return Path(f"/proc/{idle_pid}/stat").read_text().split()[2] == "Z"
            except FileNotFoundError:
                return True
        wait_for(stopped, "idle server expiry", timeout=30)
        pass_check("heap/idle defaults/expiry")
        clean_logs()
    finally:
        for workspace in workspaces:
            (workspace / "release").touch()
        for child in processes:
            if child.poll() is None:
                os.killpg(child.pid, signal.SIGTERM)
                child.wait(timeout=20)
        for index, base in enumerate(bases):
            if (base / "server/server.pid.txt").exists():
                run(["shutdown"], index=index)
        (evidence / "result.json").write_text(json.dumps({"passed": checks}, indent=2) + "\n")


if __name__ == "__main__":
    main()

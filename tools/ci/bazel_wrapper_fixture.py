"""Dependency-free rules and processes driven by check_bazel_wrapper.py."""

DEFS = '''def _probe(ctx):
    out = ctx.actions.declare_file(ctx.label.name + ".py")
    ctx.actions.symlink(output = out, target_file = ctx.file.src, is_executable = True)
    return [DefaultInfo(executable = out, runfiles = ctx.runfiles(files = ctx.files.data))]
probe = rule(implementation = _probe, executable = True, attrs = {
    "src": attr.label(allow_single_file = True),
    "data": attr.label_list(allow_files = True),
})
def _action(ctx):
    out = ctx.actions.declare_file(ctx.label.name + ".json")
    ctx.actions.run(executable = ctx.file.src, inputs = [ctx.file.src], outputs = [out],
        arguments = [out.path, ctx.attr.directory], use_default_shell_env = True,
        execution_requirements = {"no-sandbox": "1"})
    return [DefaultInfo(files = depset([out]))]
action = rule(implementation = _action, attrs = {
    "src": attr.label(allow_single_file = True), "directory": attr.string(),
})
'''

PROBE = '''#!/usr/bin/env python3
import json, os, pathlib, signal, subprocess, sys, time
mode, output, *args = sys.argv[1:]
out = pathlib.Path(output)
if mode == "nested":
    result = subprocess.run(["bazel", "--output_base=" + args[0],
        "--host_jvm_args=-Xmx256m", "--host_jvm_args=-XX:ActiveProcessorCount=2",
        "--max_idle_secs=5", "build", "--incompatible_autoload_externally=",
        "--build_event_json_file=" + output + ".bep", "//:probe"],
        cwd=os.environ["BUILD_WORKSPACE_DIRECTORY"], text=True, capture_output=True)
    out.write_text(json.dumps({"code": result.returncode, "stdout": result.stdout,
        "stderr": result.stderr}))
    sys.exit(result.returncode)
if mode == "wait":
    out.write_text(str(os.getpid()))
    signal.signal(signal.SIGTERM, lambda *_: sys.exit(42))
    signal.signal(signal.SIGINT, lambda *_: sys.exit(43))
    while True: time.sleep(.1)
env = {key: os.environ.get(key) for key in (
    "GH_TOKEN", "NEW_PROVIDER_SECRET", "DOCKER_CONFIG", "BAZEL_REAL",
    "BAZELISK_SKIP_WRAPPER", "BUILD_WORKSPACE_DIRECTORY", "BUILD_WORKING_DIRECTORY",
    "TADOKU_TEST_POSTGRES_URL", "TADOKU_TEST_PGBOUNCER_URL", "TADOKU_TEST_VALKEY_URL",
    "TADOKU_TEST_EMAIL_GATEWAY_RENDER")}
docker = pathlib.Path(os.environ.get("DOCKER_CONFIG", os.environ["HOME"] + "/.docker"))
out.write_text(json.dumps({"argv": args, "stdin": sys.stdin.read(), "env": env,
    "cwd": os.getcwd(), "runfile": pathlib.Path("payload.txt").read_text(),
    "docker": (docker / "config.json").is_file()}))
print("probe stdout")
print("probe stderr", file=sys.stderr)
sys.exit(17 if mode == "exit" else 0)
'''

ACTION = '''#!/usr/bin/env python3
import json, os, pathlib, sys, time
out, directory = map(pathlib.Path, sys.argv[1:])
directory.joinpath("started").write_text(str(time.monotonic()))
deadline = time.monotonic() + 60
while not directory.joinpath("release").exists():
    if time.monotonic() > deadline: raise SystemExit("action release timeout")
    time.sleep(.05)
out.write_text(json.dumps({key: os.environ.get(key) for key in
    ("GH_TOKEN", "NEW_PROVIDER_SECRET", "TADOKU_TEST_POSTGRES_URL")}))
directory.joinpath("finished").write_text(str(time.monotonic()))
'''

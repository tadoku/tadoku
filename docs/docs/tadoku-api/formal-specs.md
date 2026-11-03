---
title: TLA+ models
description: How Tadoku API state machines are modelled in TLA+ and checked by TLC as Bazel tests, how to run and read the checks, and how models are kept in step with the code they describe.
sidebar_position: 5.5
---

# TLA+ models

Read this when you change a TLA+ model, or change code that a model describes.

A model is a small, hand-written TLA+ description of a concurrent design. TLC, the TLA+ model checker, explores
every reachable state of the model within small bounds and fails when any
ordering of steps breaks a rule. Models catch design bugs where each line of
code is correct but two transactions interleave badly. They do not test the Go
or SQL itself, and nothing checks automatically that the code matches a model.

## Before changing code

Search `services/tadoku-api/spec/tla/` for the path of the file you are
changing. A model's header must name every source file it describes. When yours
is named, change the model in the same pull request so that it still describes
the code, and run its check.

## Run the checks

Each model has a TLC test target in `services/tadoku-api/spec/tla/BUILD.bazel`:

```sh
bazel test //services/tadoku-api/spec/tla/...              # every CI model
bazel test //services/tadoku-api/spec/tla:<model>_test     # one model
```

TLC runs on the remote JDK that `.bazelrc` selects and `MODULE.bazel` pins
through rules_java, so no local Java install is needed. Targets tagged `manual` are not part of `//...`; run them by name.

## Layout

A model is `<Name>.tla` with its CI configuration `<Name>.cfg`, which sets the
bounds and lists the invariants and properties to check. A configuration that
checks larger bounds, or a property the code is known to break, is
`<Name>_<purpose>.cfg` with a `manual` test target. Add a target with the
`tlc_test` macro from `services/tadoku-api/spec/tla/tlc.bzl`. The macro fails
when a target with a larger `size` or any `timeout` is not tagged `manual`.

## Write a model

- Write plain TLA+, not PlusCal.
- Start the module with a comment naming every source file it describes.
- Give each SQL query or Go transition one action with the same name.
- Model one SQL statement that locks its rows as one action. Split a
  transaction that takes several locks into one action per step, tracked by a
  program counter, so other transactions can run between the steps. Split a
  statement that reads before it writes without taking a lock, such as an
  `insert … select`, into a read action and a write action.
- Leave clocks out. Let a lease or deadline lapse through its own action, so
  TLC tries it at every point.
- Keep CI bounds small, such as two workers and two jobs, so each model checks
  in seconds. Put larger bounds in a `manual` configuration.
- Check a model against a deliberate bug before trusting it: change an action
  the way a plausible code bug would, confirm TLC fails on the expected
  property, then revert.

When the code breaks a property, put that property in a `manual`
configuration. Its counterexample is the reproduction for the fix, and the
fixing pull request moves the property into the CI configuration.

## CI cost

The models run in the `bazel test //...` step of `build-bazel.yaml`. Bazel
caches test results by their inputs, and a model's inputs are only its `.tla`
and `.cfg` files, the runner script and the pinned tools. A pull request that
changes no model gets cached results. One that edits a model reruns only that
model. Changing TLC, the runner script or the JDK reruns them all once. The
`small` test size stops a model after 60 seconds; keep each CI configuration
to a few seconds. TLC runs with two worker threads and a 512 MiB heap.

Because code changes never rerun a model, CI does not notice when code drifts
from its model. Keeping them in step is the author's job, as described in
[Before changing code](#before-changing-code).

## Read a counterexample

When a check fails, the test log shows TLC's error and a numbered trace. The
first state is the initial state, and each later state names the action that
produced it and the variables that changed. The trace is the shortest sequence
of steps that leads to the violation. Read it as a story, then decide whether
the code, the model or the property is wrong.

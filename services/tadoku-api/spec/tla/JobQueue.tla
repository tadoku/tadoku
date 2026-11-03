---- MODULE JobQueue ----
\* The jobs table and the worker loop that claims, runs and finishes jobs.
\*
\* Sources:
\*   services/tadoku-api/sql/jobqueue/jobqueue.sql
\*   services/tadoku-api/migrations/0031_async_outbox.up.sql
\*   services/tadoku-api/features/jobqueue/jobqueue_repository.go
\*   services/tadoku-api/features/jobqueue/jobqueue_service.go
\*   services/tadoku-api/app/worker/application.go
\*   services/tadoku-api/app/worker/runtime.go
\*
\* Each action is named after the query or worker step it models. Time is
\* abstract: a lease lapses through Expire. A successful Renew changes nothing
\* here, and a failed one is Drop. A claim's token is the job's attempt number,
\* which is unique per claim. Handlers run at least once, so effects can exceed
\* 1 and handlers must be idempotent. Operator replays are bounded by
\* MaxReplays, and replay op inserts into slot NumJobs + op. A replay is one
\* transaction: it locks the failed job, then inserts only when that job has no
\* replay yet. Every worker uses the same MaxAttempts; a rolling deploy that
\* changes it is out of scope.
EXTENDS Naturals, FiniteSets

CONSTANTS Workers, NumJobs, MaxReplays, MaxAttempts

Originals == 1..NumJobs
Ids == 1..(NumJobs + MaxReplays)
Operators == 1..MaxReplays
None == 0
Live == {"pending", "running"}
Ended == {"completed", "failed", "removed"}
Idle == [job |-> None, tok |-> None]

VARIABLES
    state,     \* jobs.state, or "absent" for an unused replay slot and "removed" after cleanup
    attempts,  \* jobs.attempts
    token,     \* jobs.claim_token as the attempt that set it, None when not running
    lease,     \* jobs.lease_expires_at is in the future
    replayOf,  \* jobs.replay_of_id
    effects,   \* committed handler runs per job
    completedAt, \* jobs.completed_at is set
    failedAt,  \* jobs.failed_at is set
    held,      \* the claim each worker holds
    phase,     \* where each worker is in runner.process
    final,     \* the finalAttemptKey value the worker gave its handler
    opPc,      \* each operator replay: "idle", "locked", "checked" or "done"
    opTarget   \* the failed job each operator replay locked

vars == <<state, attempts, token, lease, replayOf, effects, completedAt, failedAt, held, phase, final, opPc, opTarget>>

TypeOK ==
    /\ state \in [Ids -> {"absent", "pending", "running", "completed", "failed", "removed"}]
    /\ attempts \in [Ids -> 0..MaxAttempts]
    /\ token \in [Ids -> 0..MaxAttempts]
    /\ lease \in [Ids -> BOOLEAN]
    /\ replayOf \in [Ids -> {None} \cup Ids]
    /\ effects \in [Ids -> 0..MaxAttempts]
    /\ completedAt \in [Ids -> BOOLEAN]
    /\ failedAt \in [Ids -> BOOLEAN]
    /\ held \in [Workers -> [job : {None} \cup Ids, tok : 0..MaxAttempts]]
    /\ phase \in [Workers -> {"idle", "handling", "succeeded", "transient", "permanent"}]
    /\ final \in [Workers -> BOOLEAN]
    /\ opPc \in [Operators -> {"idle", "locked", "checked", "done"}]
    /\ opTarget \in [Operators -> {None} \cup Ids]

Init ==
    /\ state = [j \in Ids |-> IF j \in Originals THEN "pending" ELSE "absent"]
    /\ attempts = [j \in Ids |-> 0]
    /\ token = [j \in Ids |-> None]
    /\ lease = [j \in Ids |-> FALSE]
    /\ replayOf = [j \in Ids |-> None]
    /\ effects = [j \in Ids |-> 0]
    /\ completedAt = [j \in Ids |-> FALSE]
    /\ failedAt = [j \in Ids |-> FALSE]
    /\ held = [w \in Workers |-> Idle]
    /\ phase = [w \in Workers |-> "idle"]
    /\ final = [w \in Workers |-> FALSE]
    /\ opPc = [op \in Operators |-> "idle"]
    /\ opTarget = [op \in Operators |-> None]

\* The guard every fenced query shares: the claim token matches, the job is
\* running and its lease has not expired.
Holds(w) ==
    LET j == held[w].job IN
    /\ j # None
    /\ state[j] = "running"
    /\ token[j] = held[w].tok
    /\ lease[j]

Release(w) ==
    /\ held' = [held EXCEPT ![w] = Idle]
    /\ phase' = [phase EXCEPT ![w] = "idle"]
    /\ final' = [final EXCEPT ![w] = FALSE]

\* Claim: a pending job, or a running job whose lease expired, below MaxAttempts.
Claim(w, j) ==
    /\ phase[w] = "idle"
    /\ attempts[j] < MaxAttempts
    /\ \/ state[j] = "pending"
       \/ state[j] = "running" /\ ~lease[j]
    /\ state' = [state EXCEPT ![j] = "running"]
    /\ attempts' = [attempts EXCEPT ![j] = @ + 1]
    /\ token' = [token EXCEPT ![j] = attempts[j] + 1]
    /\ lease' = [lease EXCEPT ![j] = TRUE]
    /\ held' = [held EXCEPT ![w] = [job |-> j, tok |-> attempts[j] + 1]]
    /\ phase' = [phase EXCEPT ![w] = "handling"]
    /\ final' = [final EXCEPT ![w] = attempts[j] + 1 >= MaxAttempts]
    /\ UNCHANGED <<replayOf, effects, completedAt, failedAt, opPc, opTarget>>

\* The lease runs out before the holder renews it.
Expire(j) ==
    /\ state[j] = "running"
    /\ lease[j]
    /\ lease' = [lease EXCEPT ![j] = FALSE]
    /\ UNCHANGED <<state, attempts, token, replayOf, effects, completedAt, failedAt, held, phase, final, opPc, opTarget>>

\* The handler commits its own transaction, separate from the job transition.
HandlerSucceeds(w) ==
    /\ phase[w] = "handling"
    /\ effects' = [effects EXCEPT ![held[w].job] = @ + 1]
    /\ phase' = [phase EXCEPT ![w] = "succeeded"]
    /\ UNCHANGED <<state, attempts, token, lease, replayOf, completedAt, failedAt, held, final, opPc, opTarget>>

HandlerFails(w, kind) ==
    /\ phase[w] = "handling"
    /\ phase' = [phase EXCEPT ![w] = kind]
    /\ UNCHANGED <<state, attempts, token, lease, replayOf, effects, completedAt, failedAt, held, final, opPc, opTarget>>

Complete(w) ==
    LET j == held[w].job IN
    /\ phase[w] = "succeeded"
    /\ Holds(w)
    /\ state' = [state EXCEPT ![j] = "completed"]
    /\ token' = [token EXCEPT ![j] = None]
    /\ lease' = [lease EXCEPT ![j] = FALSE]
    /\ completedAt' = [completedAt EXCEPT ![j] = TRUE]
    /\ Release(w)
    /\ UNCHANGED <<attempts, replayOf, effects, failedAt, opPc, opTarget>>

\* Retry: back to pending, or failed once attempts reach MaxAttempts. A handler
\* that returned nil after its deadline counts as a transient error, so a
\* committed effect can still be retried.
Retry(w) ==
    LET j == held[w].job IN
    /\ phase[w] \in {"transient", "succeeded"}
    /\ Holds(w)
    /\ state' = [state EXCEPT ![j] = IF attempts[j] >= MaxAttempts THEN "failed" ELSE "pending"]
    /\ token' = [token EXCEPT ![j] = None]
    /\ lease' = [lease EXCEPT ![j] = FALSE]
    /\ failedAt' = [failedAt EXCEPT ![j] = attempts[j] >= MaxAttempts]
    /\ Release(w)
    /\ UNCHANGED <<attempts, replayOf, effects, completedAt, opPc, opTarget>>

\* Fail: a permanent handler error.
Fail(w) ==
    LET j == held[w].job IN
    /\ phase[w] = "permanent"
    /\ Holds(w)
    /\ state' = [state EXCEPT ![j] = "failed"]
    /\ token' = [token EXCEPT ![j] = None]
    /\ lease' = [lease EXCEPT ![j] = FALSE]
    /\ failedAt' = [failedAt EXCEPT ![j] = TRUE]
    /\ Release(w)
    /\ UNCHANGED <<attempts, replayOf, effects, completedAt, opPc, opTarget>>

\* A Renew or a transition finds the lease lost, and the worker gives up its claim.
Drop(w) ==
    /\ phase[w] # "idle"
    /\ ~Holds(w)
    /\ Release(w)
    /\ UNCHANGED <<state, attempts, token, lease, replayOf, effects, completedAt, failedAt, opPc, opTarget>>

\* The worker stops mid-job without a transition: a crash, a failed transition
\* query, the claim found too close to expiry before dispatch, or the renew loop
\* cancelling the handler before its margin runs out while the lease is live.
Crash(w) ==
    /\ phase[w] # "idle"
    /\ Release(w)
    /\ UNCHANGED <<state, attempts, token, lease, replayOf, effects, completedAt, failedAt, opPc, opTarget>>

\* FailExhausted: a due or lease-expired job that has used every attempt.
FailExhausted(j) ==
    /\ attempts[j] >= MaxAttempts
    /\ \/ state[j] = "pending"
       \/ state[j] = "running" /\ ~lease[j]
    /\ state' = [state EXCEPT ![j] = "failed"]
    /\ token' = [token EXCEPT ![j] = None]
    /\ lease' = [lease EXCEPT ![j] = FALSE]
    /\ failedAt' = [failedAt EXCEPT ![j] = TRUE]
    /\ UNCHANGED <<attempts, replayOf, effects, completedAt, held, phase, final, opPc, opTarget>>

\* CleanupCompleted: a completed job that is not a replay and that no replay
\* refers to. Keeping completed replays keeps their lineage unreplayable.
CleanupCompleted(j) ==
    /\ state[j] = "completed"
    /\ replayOf[j] = None
    /\ \A k \in Ids : replayOf[k] # j
    /\ state' = [state EXCEPT ![j] = "removed"]
    /\ UNCHANGED <<attempts, token, lease, replayOf, effects, completedAt, failedAt, held, phase, final, opPc, opTarget>>

\* An operator replay holds the failed job's row lock until it commits or
\* refuses.
LockHolder(op, j) == opPc[op] \in {"locked", "checked"} /\ opTarget[op] = j

\* LockReplaySource: select ... for update on a failed job of a supported type.
\* It waits while another replay holds the lock; a job that is not failed is
\* refused without a step.
ReplayLock(op, j) ==
    /\ opPc[op] = "idle"
    /\ state[j] = "failed"
    /\ \A other \in Operators : ~LockHolder(other, j)
    /\ opPc' = [opPc EXCEPT ![op] = "locked"]
    /\ opTarget' = [opTarget EXCEPT ![op] = j]
    /\ UNCHANGED <<state, attempts, token, lease, replayOf, effects, completedAt, failedAt, held, phase, final>>

\* Replay is an insert ... select guarded by "no replay of this job exists",
\* so it reads before it writes. A removed row no longer exists.
ReplayCheck(op) ==
    LET replayed ==
            \E k \in Ids : replayOf[k] = opTarget[op] /\ state[k] \notin {"absent", "removed"}
    IN
    /\ opPc[op] = "locked"
    /\ opPc' = [opPc EXCEPT ![op] = IF replayed THEN "done" ELSE "checked"]
    /\ UNCHANGED <<state, attempts, token, lease, replayOf, effects, completedAt, failedAt, held, phase, final, opTarget>>

ReplayInsert(op) ==
    LET slot == NumJobs + op IN
    /\ opPc[op] = "checked"
    /\ state' = [state EXCEPT ![slot] = "pending"]
    /\ replayOf' = [replayOf EXCEPT ![slot] = opTarget[op]]
    /\ opPc' = [opPc EXCEPT ![op] = "done"]
    /\ UNCHANGED <<attempts, token, lease, effects, completedAt, failedAt, held, phase, final, opTarget>>

WorkerStep(w) ==
    \/ \E j \in Ids : Claim(w, j)
    \/ HandlerSucceeds(w)
    \/ HandlerFails(w, "transient")
    \/ HandlerFails(w, "permanent")
    \/ Complete(w)
    \/ Retry(w)
    \/ Fail(w)
    \/ Drop(w)

\* Every job has ended and every worker is idle. Only this state may stutter,
\* so TLC still reports any other state with no next step as a deadlock.
Done ==
    /\ \A j \in Ids : state[j] \notin Live
    /\ \A w \in Workers : phase[w] = "idle"
    /\ UNCHANGED vars

Next ==
    \/ Done
    \/ \E w \in Workers : WorkerStep(w) \/ Crash(w)
    \/ \E j \in Ids : Expire(j) \/ FailExhausted(j) \/ CleanupCompleted(j)
    \/ \E op \in Operators : ReplayCheck(op) \/ ReplayInsert(op) \/ \E j \in Ids : ReplayLock(op, j)

\* Workers keep working, leases eventually lapse, the exhaustion sweep runs and
\* a started replay commits. Crashes, cleanup and starting an operator replay
\* are never forced. This assumes that a
\* handler returns soon after its context is cancelled: the worker waits for
\* dispatch to return before its slot is free, so a handler that ignores
\* cancellation keeps the slot, and Drop releasing it at once would hide that.
\* It also assumes every job's type is registered and its tenant is served.
Fairness ==
    /\ \A w \in Workers : WF_vars(WorkerStep(w))
    /\ \A j \in Ids : WF_vars(Expire(j)) /\ WF_vars(FailExhausted(j))
    /\ \A op \in Operators : WF_vars(ReplayCheck(op)) /\ WF_vars(ReplayInsert(op))

Spec == Init /\ [][Next]_vars /\ Fairness

----
\* Safety

\* jobs_claim_running: a running job has a claim token, and only a running job has a lease.
ClaimRunning ==
    \A j \in Ids :
        /\ (state[j] = "running") = (token[j] # None)
        /\ lease[j] => state[j] = "running"

AttemptsBounded == \A j \in Ids : attempts[j] <= MaxAttempts

\* jobs_terminal_timestamps: completed_at and failed_at match the state.
TerminalTimestamps ==
    \A j \in Ids :
        state[j] \in {"pending", "running", "completed", "failed"} =>
            /\ completedAt[j] = (state[j] = "completed")
            /\ failedAt[j] = (state[j] = "failed")

\* A sanity check: a job only completes after its handler committed.
CompletedHadEffect == \A j \in Ids : state[j] \in {"completed", "removed"} => effects[j] >= 1

\* Completed and failed never change, except that cleanup removes completed jobs.
TerminalIsFinal ==
    [][\A j \in Ids :
        \/ state[j] \notin {"completed", "failed", "removed"}
        \/ state'[j] = state[j]
        \/ state[j] = "completed" /\ state'[j] = "removed"]_vars

\* While a lease is live, only the worker holding it can end or replace the claim.
LeaseRespected ==
    [][\A j \in Ids :
        (lease[j] /\ token'[j] # token[j]) =>
            \E w \in Workers : held[w].job = j /\ held[w].tok = token[j] /\ held'[w] = Idle]_vars

\* A Retry fails the job exactly when its handler was told this was the final
\* attempt. A job can still exhaust its attempts without a final-attempt run,
\* when a later claim never reached its handler.
FinalAttemptAgrees ==
    [][\A w \in Workers :
        LET j == held[w].job IN
        (/\ phase[w] \in {"transient", "succeeded"}
         /\ j # None
         /\ held'[w] = Idle
         /\ state[j] = "running"
         /\ state'[j] \in {"pending", "failed"}) =>
            ((state'[j] = "failed") = final[w])]_vars

\* Liveness: every job that exists eventually completes or fails.
EveryJobEnds == \A j \in Ids : (state[j] \in Live) ~> (state[j] \in Ended)

----
\* Replay rules. A replay is "unfinished" while it is pending or running.

RECURSIVE Root(_)
Root(j) == IF replayOf[j] = None THEN j ELSE Root(replayOf[j])

Lineage(r) == {j \in Ids : state[j] # "absent" /\ Root(j) = r}

\* At most one job per lineage is pending or running.
NoConcurrentReplays ==
    \A r \in Originals : Cardinality({j \in Lineage(r) : state[j] \in Live}) <= 1

\* Once a job in a lineage completes, nothing in that lineage is pending or
\* running. A removed job had completed, so cleanup does not erase the success.
NoReplayAfterSuccess ==
    \A r \in Originals :
        (\E j \in Lineage(r) : state[j] \in {"completed", "removed"}) =>
            \A j \in Lineage(r) : state[j] \notin Live
====

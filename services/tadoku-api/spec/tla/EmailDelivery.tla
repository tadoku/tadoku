---- MODULE EmailDelivery ----
\* Delivering one retained email through its email.send.v1 jobs, against the
\* worker maintenance that settles emails whose jobs failed and operator replays.
\*
\* Sources:
\*   services/tadoku-api/features/email/email_service.go
\*   services/tadoku-api/features/email/email_repository.go
\*   services/tadoku-api/sql/email/email.sql
\*   services/tadoku-api/migrations/0052_identity_v2.up.sql
\*   services/tadoku-api/app/worker/application.go
\*   services/tadoku-api/app/worker/emails.go
\*   services/tadoku-api/app/worker/runtime.go
\*   services/tadoku-api/infra/smtp/smtp.go
\*   services/tadoku-api/sql/jobqueue/jobqueue.sql
\*   services/tadoku-api/features/jobqueue/jobqueue_repository.go
\*   services/tadoku-api/features/jobqueue/jobqueue_service.go
\*   services/tadoku-api/spec/tla/JobQueue.tla (the job contract)
\*
\* The job queue is JobQueue.tla's contract, not re-modelled: a claim is fenced
\* by its token and lease, every claim is an attempt, a retry at MaxAttempts
\* fails the job, FailExhausted fails an exhausted job whose lease lapsed,
\* cleanup removes only a completed job that is not a replay and has none, and
\* a replay inserts one pending job only while every job in the lineage has
\* failed and the target has no replay. JobQueue.tla checks that guard under
\* its lineage lock, so a replay is one step here. Replay op inserts slot 1 + op
\* and replays the slot before it: under the guard, a failed job without a
\* replay is always the lineage's last one. Lineages replayed before the guard,
\* JobQueue.tla's ForkedHistory, are out of scope.
\*
\* SendAttempt runs its statements outside any transaction: Find, then one
\* Finish update. It sends a pending or failed email, so a replay resends a
\* failed one. Finish(held) replaces only a pending row; Finish(sent) and
\* Finish(failed) also replace a failed one. Finish runs on a
\* detached context and is not fenced by the claim, so a handler whose lease
\* lapsed still records its outcome; two workers let such a stale run overlap a
\* fresh claim. A run can crash at any step, a Finish statement can fail, and the
\* SMTP exchange can fail after the provider accepted the message. A handler
\* that returns nil after its deadline is retried. ReconcileFailures is one
\* update: its jobs subqueries read the statement's snapshot, and the emails
\* row is rechecked for status = 'pending' when the update locks it.
\* Leaving clocks out lets a stale Finish land at any time. In the code it lands
\* after the lease lapsed only with clock skew, a starved renew loop or a commit
\* that outlives its cancelled client; a replay inserted after reconciliation's
\* snapshot needs none of these.
\* SendAttempt and ReconcileFailures share the hold rule, fixed per tenant and
\* sender, so both values are checked. Retention deletes rows only after days
\* and is not modelled; nor are error texts.
EXTENDS Naturals, FiniteSets

CONSTANTS Workers, MaxAttempts, MaxReplays

Ids == 1..(1 + MaxReplays)
Operators == 1..MaxReplays
None == 0
Live == {"pending", "running"}
Gone == {"absent", "removed"}
Idle == [job |-> None, tok |-> None, pc |-> "idle", final |-> FALSE]

VARIABLES
    email,     \* emails.status
    delivered, \* messages the provider accepted
    hold,      \* the hold rule: a non-production tenant or no sender
    lostSent,  \* a Finish(sent) after delivery left the email failed
    state,     \* jobs.state, "absent" for an unused replay slot and "removed" after cleanup
    attempts,  \* jobs.attempts
    token,     \* jobs.claim_token as the attempt that set it, None when not running
    lease,     \* jobs.lease_expires_at is in the future
    run,       \* each worker's claim and where its SendAttempt is
    recon      \* ReconcileFailures: "idle", or "write" once its snapshot found only failed jobs

vars == <<email, delivered, hold, lostSent, state, attempts, token, lease, run, recon>>

MaxDeliveries == Cardinality(Ids) * MaxAttempts

TypeOK ==
    /\ email \in {"pending", "sent", "held", "failed"}
    /\ delivered \in 0..MaxDeliveries
    /\ hold \in BOOLEAN
    /\ lostSent \in BOOLEAN
    /\ state \in [Ids -> {"absent", "pending", "running", "completed", "failed", "removed"}]
    /\ attempts \in [Ids -> 0..MaxAttempts]
    /\ token \in [Ids -> 0..MaxAttempts]
    /\ lease \in [Ids -> BOOLEAN]
    /\ run \in [Workers -> [job : {None} \cup Ids, tok : 0..MaxAttempts,
                            pc : {"idle", "find", "send", "finishSent", "finishFailed", "finishHeld", "ok", "err"},
                            final : BOOLEAN]]
    /\ recon \in {"idle", "write"}

\* Write inserts the pending email and its job in one business transaction.
Init ==
    /\ email = "pending"
    /\ delivered = 0
    /\ hold \in BOOLEAN
    /\ lostSent = FALSE
    /\ state = [j \in Ids |-> IF j = 1 THEN "pending" ELSE "absent"]
    /\ attempts = [j \in Ids |-> 0]
    /\ token = [j \in Ids |-> None]
    /\ lease = [j \in Ids |-> FALSE]
    /\ run = [w \in Workers |-> Idle]
    /\ recon = "idle"

JobVars == <<state, attempts, token, lease>>

Goto(w, pc) == run' = [run EXCEPT ![w].pc = pc]

----
\* The job queue, as JobQueue.tla checks it.

Holds(w) ==
    LET j == run[w].job IN
    /\ j # None
    /\ state[j] = "running"
    /\ token[j] = run[w].tok
    /\ lease[j]

Claim(w, j) ==
    /\ run[w].pc = "idle"
    /\ attempts[j] < MaxAttempts
    /\ \/ state[j] = "pending"
       \/ state[j] = "running" /\ ~lease[j]
    /\ state' = [state EXCEPT ![j] = "running"]
    /\ attempts' = [attempts EXCEPT ![j] = @ + 1]
    /\ token' = [token EXCEPT ![j] = attempts[j] + 1]
    /\ lease' = [lease EXCEPT ![j] = TRUE]
    /\ run' = [run EXCEPT ![w] = [job |-> j, tok |-> attempts[j] + 1, pc |-> "find",
                                  final |-> attempts[j] + 1 >= MaxAttempts]]
    /\ UNCHANGED <<email, delivered, hold, lostSent, recon>>

Expire(j) ==
    /\ state[j] = "running"
    /\ lease[j]
    /\ lease' = [lease EXCEPT ![j] = FALSE]
    /\ UNCHANGED <<email, delivered, hold, lostSent, state, attempts, token, run, recon>>

End(w, next) ==
    LET j == run[w].job IN
    /\ Holds(w)
    /\ state' = [state EXCEPT ![j] = next]
    /\ token' = [token EXCEPT ![j] = None]
    /\ lease' = [lease EXCEPT ![j] = FALSE]
    /\ run' = [run EXCEPT ![w] = Idle]
    /\ UNCHANGED <<email, delivered, hold, lostSent, attempts, recon>>

\* The handler returned nil within its deadline.
Complete(w) == run[w].pc = "ok" /\ End(w, "completed")

\* The handler returned an error, or nil after its deadline.
Retry(w) ==
    /\ run[w].pc \in {"ok", "err"}
    /\ End(w, IF attempts[run[w].job] >= MaxAttempts THEN "failed" ELSE "pending")

\* The renew loop or the transition finds the lease lost and records nothing.
Drop(w) ==
    /\ run[w].pc \in {"ok", "err"}
    /\ ~Holds(w)
    /\ run' = [run EXCEPT ![w] = Idle]
    /\ UNCHANGED <<email, delivered, hold, lostSent, state, attempts, token, lease, recon>>

\* The worker stops at any step without recording an outcome.
Crash(w) ==
    /\ run[w].pc # "idle"
    /\ run' = [run EXCEPT ![w] = Idle]
    /\ UNCHANGED <<email, delivered, hold, lostSent, state, attempts, token, lease, recon>>

FailExhausted(j) ==
    /\ attempts[j] >= MaxAttempts
    /\ \/ state[j] = "pending"
       \/ state[j] = "running" /\ ~lease[j]
    /\ state' = [state EXCEPT ![j] = "failed"]
    /\ token' = [token EXCEPT ![j] = None]
    /\ lease' = [lease EXCEPT ![j] = FALSE]
    /\ UNCHANGED <<email, delivered, hold, lostSent, attempts, run, recon>>

CleanupCompleted ==
    /\ state[1] = "completed"
    /\ MaxReplays > 0 => state[2] = "absent"
    /\ state' = [state EXCEPT ![1] = "removed"]
    /\ UNCHANGED <<email, delivered, hold, lostSent, attempts, token, lease, run, recon>>

\* tadoku-worker replay of the lineage's failed job without a replay.
Replay(op) ==
    LET slot == 1 + op IN
    /\ state[slot] = "absent"
    /\ state[slot - 1] = "failed"
    /\ \A k \in Ids : state[k] \notin Live \cup {"completed", "removed"}
    /\ state' = [state EXCEPT ![slot] = "pending"]
    /\ UNCHANGED <<email, delivered, hold, lostSent, attempts, token, lease, run, recon>>

----
\* SendAttempt, one step per statement or provider exchange.

\* Find: a missing, sent or held email returns nil; a held tenant finishes held;
\* a pending or failed email is sent.
Find(w) ==
    /\ run[w].pc = "find"
    /\ Goto(w, CASE email \in {"sent", "held"} -> "ok"
                 [] hold -> "finishHeld"
                 [] OTHER -> "send")
    /\ UNCHANGED <<email, delivered, hold, lostSent, state, attempts, token, lease, recon>>

\* Sender.Send: the provider accepts and the reply arrives, the exchange fails
\* before acceptance, or it fails after acceptance. Only the final attempt
\* records a failure.
SendAccepted(w) ==
    /\ run[w].pc = "send"
    /\ delivered' = delivered + 1
    /\ Goto(w, "finishSent")
    /\ UNCHANGED <<email, hold, lostSent, state, attempts, token, lease, recon>>

SendFails(w) ==
    /\ run[w].pc = "send"
    /\ \E accepted \in BOOLEAN : delivered' = IF accepted THEN delivered + 1 ELSE delivered
    /\ Goto(w, IF run[w].final THEN "finishFailed" ELSE "err")
    /\ UNCHANGED <<email, hold, lostSent, state, attempts, token, lease, recon>>

\* Finish: update emails ... where status = 'pending', or 'failed' for sent and failed.
FinishTo(w, from, status, replaces, next) ==
    /\ run[w].pc = from
    /\ email' = IF email \in replaces THEN status ELSE email
    /\ Goto(w, next)
    /\ UNCHANGED <<delivered, hold, state, attempts, token, lease, recon>>

FinishSent(w) ==
    /\ FinishTo(w, "finishSent", "sent", {"pending", "failed"}, "ok")
    /\ lostSent' = (lostSent \/ email' = "failed")

FinishFailed(w) == FinishTo(w, "finishFailed", "failed", {"pending", "failed"}, "err") /\ UNCHANGED lostSent

FinishHeld(w) == FinishTo(w, "finishHeld", "held", {"pending"}, "ok") /\ UNCHANGED lostSent

\* The Finish statement fails and the handler returns its error.
FinishErrors(w) ==
    /\ run[w].pc \in {"finishSent", "finishFailed", "finishHeld"}
    /\ Goto(w, "err")
    /\ UNCHANGED <<email, delivered, hold, lostSent, state, attempts, token, lease, recon>>

----
\* ReconcileFailures, from the worker's maintenance pass.

\* The statement's snapshot finds the pending email with a failed job and no
\* job in another state; a removed job no longer exists.
ReconcileRead ==
    /\ recon = "idle"
    /\ email = "pending"
    /\ \E j \in Ids : state[j] = "failed"
    /\ \A j \in Ids : state[j] \in Gone \cup {"failed"}
    /\ recon' = "write"
    /\ UNCHANGED <<email, delivered, hold, lostSent, state, attempts, token, lease, run>>

\* The update locks the row and rechecks only its status.
ReconcileWrite ==
    /\ recon = "write"
    /\ email' = IF email = "pending" THEN IF hold THEN "held" ELSE "failed" ELSE email
    /\ recon' = "idle"
    /\ UNCHANGED <<delivered, hold, lostSent, state, attempts, token, lease, run>>

----

WorkerStep(w) ==
    \/ \E j \in Ids : Claim(w, j)
    \/ Find(w) \/ SendAccepted(w) \/ SendFails(w)
    \/ FinishSent(w) \/ FinishFailed(w) \/ FinishHeld(w) \/ FinishErrors(w)
    \/ Complete(w) \/ Retry(w) \/ Drop(w)

\* Every job has ended, every worker is idle and reconciliation is between passes.
Done ==
    /\ \A j \in Ids : state[j] \notin Live
    /\ \A w \in Workers : run[w].pc = "idle"
    /\ recon = "idle"
    /\ UNCHANGED vars

Next ==
    \/ Done
    \/ \E w \in Workers : WorkerStep(w) \/ Crash(w)
    \/ \E j \in Ids : Expire(j) \/ FailExhausted(j)
    \/ CleanupCompleted
    \/ \E op \in Operators : Replay(op)
    \/ ReconcileRead \/ ReconcileWrite

\* Workers keep working, leases lapse, the exhaustion and reconciliation passes
\* run. Crashes, cleanup and replays are never forced. As in JobQueue.tla, the
\* job's tenant is served by a worker that registers email.send.v1.
Fairness ==
    /\ \A w \in Workers : WF_vars(WorkerStep(w))
    /\ \A j \in Ids : WF_vars(Expire(j)) /\ WF_vars(FailExhausted(j))
    /\ WF_vars(ReconcileRead) /\ WF_vars(ReconcileWrite)

Spec == Init /\ [][Next]_vars /\ Fairness

----
\* Safety

SentHadDelivery == email = "sent" => delivered >= 1

\* A held tenant never reaches the provider, and only a held tenant holds.
HoldRule ==
    /\ hold => delivered = 0 /\ email \in {"pending", "held"}
    /\ ~hold => email # "held"

\* A send job completes only once its email has left pending.
CompletedJobNotPending == \A j \in Ids : state[j] \in {"completed", "removed"} => email # "pending"

\* A pending email always has a job that can still settle it.
PendingHasJob == email = "pending" => \E j \in Ids : state[j] \notin Gone

\* A settled email never changes, except that a Finish(sent) turns failed into sent.
\* A newer failure leaves it failed; error texts are not modelled.
TerminalEmailFinal ==
    [][email # "pending" =>
        \/ email' = email
        \/ /\ email = "failed" /\ email' = "sent"
           /\ \E w \in Workers : run[w].pc = "finishSent" /\ run'[w].pc = "ok"]_vars

\* A delivery whose Finish(sent) runs is recorded as sent: it fails when Finish(sent)
\* does not replace a failed row.
SentNeverLost == ~lostSent

\* A completed send job, such as a replay, leaves its email sent or held.
CompletedJobDelivered == \A j \in Ids : state[j] = "completed" => email \in {"sent", "held"}

\* Liveness: every email leaves pending.
EmailSettles == email = "pending" ~> email # "pending"

----
\* Accepted at-least-once behaviour, checked by a manual configuration.

\* Broken by design: the provider can receive an email twice.
AtMostOnceDelivery == delivered <= 1
====

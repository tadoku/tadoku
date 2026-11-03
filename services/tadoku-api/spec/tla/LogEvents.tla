---- MODULE LogEvents ----
\* One log's changes and the contest events they publish, against every request
\* and job that changes the same log or its contest attachments.
\*
\* Sources:
\*   services/tadoku-api/app/api/log_mutations.go
\*   services/tadoku-api/app/api/log_attachments.go
\*   services/tadoku-api/app/api/contests.go (UpsertContestRegistration)
\*   services/tadoku-api/app/worker/account_deletion.go
\*   services/tadoku-api/features/logs/logs_mutation_service.go
\*   services/tadoku-api/features/logs/logs_contest_service.go
\*   services/tadoku-api/features/logs/logs_removal_service.go
\*   services/tadoku-api/features/logs/logs_service.go (FindLog)
\*   services/tadoku-api/features/contests/contests_service.go (SelectRegistrationsForScoring, ApplyRegistration)
\*   services/tadoku-api/features/contestevents/contestevents_service.go
\*   services/tadoku-api/features/profile/profile_service.go (LockUser, Tombstone)
\*   services/tadoku-api/domain/contestwindow/window.go
\*   services/tadoku-api/sql/logs/logs.sql
\*   services/tadoku-api/sql/logs/contest_logs.sql
\*   services/tadoku-api/sql/logs/user_removal.sql
\*   services/tadoku-api/sql/contests/contests.sql (DetachContestLogsForLanguages)
\*   services/tadoku-api/sql/contestevents/contestevents.sql
\*   services/tadoku-api/sql/leaderboard/leaderboard.sql (OfficialLeaderboard)
\*   docs/docs/tadoku-api/contest-events.md
\*
\* Every writer of a user's logs, contest_logs rows and registrations first
\* locks that user's users row through profile.LockUser or Tombstone and holds
\* it until it commits. So the statements and Go checks a transaction runs while
\* it holds the lock are merged into one commit step: no competing writer can run
\* between them, and each later statement reads what earlier lock holders
\* committed. AccountDeletion.tla covers the lock itself and filing in detail.
\* What the lock does not order is the clock. The events of a transaction carry
\* the time its request read with timex.Now(), and that read is its own step:
\* every request reads it after it takes the lock, and only the delete job reads
\* it before. The same reading judges contest windows for that request, except
\* that SelectRegistrationsForScoring reads the clock again for itself. A contest
\* finishes through its own step, so TLC tries every point. The clock is one
\* counter that every process shares: the model assumes the API and worker
\* clocks agree and never tie, so within one contest a log's events carry
\* distinct times and the event ID tie-break never applies. It orders
\* timestamps; nothing waits for it.
\* The FindLog a request runs before its transaction is its own step: a request
\* about a log starts only once that log is committed and not soft-deleted.
\* An update is assumed to change a field every contest sees, so it publishes
\* an update event wherever the log is visible. The code publishes nothing for
\* an unchanged view; dropping update events keeps a valid sequence valid.
\* The feed is read in occurrence order, oldest first.
EXTENDS Naturals

CONSTANTS
    Contests, \* the contests the user is registered for, covering the log's language
    Official, \* the official ones; contests.official never changes
    Actors    \* the requests and jobs this configuration runs

None == "none"
Procs == {"create", "update", "delete", "attach", "moderate", "register", "file", "deleter"}
Event == {"log.submitted", "log.updated", "log.withdrawn"}

VARIABLES
    row,        \* the logs row: absent before CreateLog, present, or purged
    deleted,    \* logs.deleted_at is set
    frozen,     \* logs.frozen_at is set
    attached,   \* contests with a contest_logs row for the log
    eligible,   \* logs.eligible_official_leaderboard
    wroteFrozen, \* a write reached the log or its contest rows while it was frozen
    described,  \* the log has a description or tags
    user,       \* active, locked (deletion_locked_at) or deleted (deleted_at)
    registered, \* contests whose registration still covers the log's language
    open,       \* contests that are unfinished at the current time
    clock,      \* the wall clock in ticks
    feed,       \* committed contest_events rows
    lockU,      \* who holds the users row lock
    pc,         \* each process's next step
    now,        \* each process's timex.Now() reading, its events' occurred_at
    seen,       \* the contests unfinished at that reading
    picked      \* the registrations SelectRegistrationsForScoring picked for an attachment change

vars == <<row, deleted, frozen, attached, eligible, wroteFrozen, described, user, registered, open, clock,
          feed, lockU, pc, now, seen, picked>>

TypeOK ==
    /\ row \in {"absent", "present", "purged"}
    /\ deleted \in BOOLEAN /\ frozen \in BOOLEAN /\ eligible \in BOOLEAN
    /\ attached \subseteq Contests /\ registered \subseteq Contests /\ open \subseteq Contests
    /\ wroteFrozen \in BOOLEAN /\ described \in BOOLEAN /\ clock \in Nat
    /\ user \in {"active", "locked", "deleted"}
    /\ feed \subseteq [c : Contests, ev : Event, t : Nat]
    /\ lockU \in Procs \cup {None}
    /\ now \in [Procs -> Nat] /\ seen \in [Procs -> SUBSET Contests] /\ picked \in [Procs -> SUBSET Contests]

Init ==
    /\ row = "absent" /\ deleted = FALSE /\ frozen = FALSE
    /\ attached = {} /\ eligible = FALSE /\ wroteFrozen = FALSE /\ described = FALSE
    /\ user = "active" /\ registered = Contests /\ open = Contests
    /\ clock = 0 /\ feed = {} /\ lockU = None
    /\ pc = [p \in Procs |-> IF p \in Actors THEN "start" ELSE "done"]
    /\ now = [p \in Procs |-> 0] /\ seen = [p \in Procs |-> {}] /\ picked = [p \in Procs |-> {}]

Goto(p, step) == pc' = [pc EXCEPT ![p] = step]

\* Commit or roll back: release the users row lock.
Finish(p) == Goto(p, "done") /\ lockU' = None

\* ReadContestLogViews: the contests the log is visible in, a frozen log included.
Visible == row = "present" /\ ~deleted
Views == IF Visible THEN attached ELSE {}

\* LockLogForMutation and lockMutableLog.
Mutable == Visible /\ ~frozen

\* Any write to the log or its contest rows.
Write == wroteFrozen' = (wroteFrozen \/ frozen)

Events(cs, ev, t) == {[c |-> c, ev |-> ev, t |-> t] : c \in cs}

\* contestLogChanges diffs the views before and after a transaction's writes.
Diff(before, after, changed, t) ==
    Events(after \ before, "log.submitted", t)
    \cup Events(before \ after, "log.withdrawn", t)
    \cup (IF changed THEN Events(before \cap after, "log.updated", t) ELSE {})

Data == <<row, deleted, frozen, attached, eligible, wroteFrozen, described, user, registered, feed>>

\* timex.Now(): the time a request stamps on its events and judges contest windows by.
Now(p, from, next) ==
    /\ pc[p] = from
    /\ clock' = clock + 1
    /\ now' = [now EXCEPT ![p] = clock + 1]
    /\ seen' = [seen EXCEPT ![p] = open]
    /\ Goto(p, next)
    /\ UNCHANGED <<Data, open, lockU, picked>>

\* profile.LockUser: a user whose deletion is filed or done refuses every write.
LockUser(p, from, next) ==
    /\ pc[p] = from
    /\ lockU = None
    /\ IF user = "active"
       THEN lockU' = p /\ Goto(p, next)
       ELSE Goto(p, "done") /\ UNCHANGED lockU
    /\ UNCHANGED <<Data, open, clock, now, seen, picked>>

\* FindLog before the transaction, which hides a soft-deleted log.
FindLog(p, next) ==
    /\ pc[p] = "start"
    /\ Goto(p, IF Visible THEN next ELSE "done")
    /\ UNCHANGED <<Data, open, clock, lockU, now, seen, picked>>

\* A transaction that writes nothing and publishes nothing.
Abort(p) == Finish(p) /\ UNCHANGED Data

TimeUnchanged == UNCHANGED <<open, clock, now, seen, picked>>

----
\* CreateLog reads the time after LockUser. SelectRegistrationsForScoring
\* picks registrations in contests unfinished when it runs; Create inserts the log,
\* its contest_logs rows and its tags, and publishes a submitted event per contest.
CreateCommit ==
    /\ pc["create"] = "commit"
    /\ \E targets \in SUBSET (registered \cap open) :
        /\ row' = "present"
        /\ attached' = targets
        /\ eligible' = (targets \cap Official # {})
        /\ feed' = feed \cup Events(targets, "log.submitted", now["create"])
    /\ described' \in BOOLEAN
    /\ Write
    /\ Finish("create")
    /\ UNCHANGED <<deleted, frozen, user, registered>>
    /\ TimeUnchanged

Create == LockUser("create", "start", "now") \/ Now("create", "now", "commit") \/ CreateCommit

\* UpdateLog reads the time after LockUser. Update: lockMutableLog, then
\* UpdateLog, UpdateOngoingContestLogs for contests unfinished at that time, the
\* tags, and the diff of the views.
UpdateCommit ==
    /\ pc["update"] = "commit"
    /\ IF Mutable
       THEN /\ Write
            /\ described' \in BOOLEAN
            /\ feed' = feed \cup Diff(Views, Views, TRUE, now["update"])
            /\ Finish("update")
            /\ UNCHANGED <<row, deleted, frozen, attached, eligible, user, registered>>
       ELSE Abort("update")
    /\ TimeUnchanged

Update ==
    FindLog("update", "lock") \/ LockUser("update", "lock", "now") \/ Now("update", "now", "commit") \/ UpdateCommit

\* DeleteLog reads the time after LockUser. Delete: lockMutableLog,
\* ListAttachedContests, refuse a log in a finished contest, then SoftDeleteLog,
\* which keeps the contest_logs rows, and a withdrawn event per contest.
DeleteCommit ==
    /\ pc["delete"] = "commit"
    /\ IF Mutable /\ attached \subseteq seen["delete"]
       THEN /\ deleted' = TRUE
            /\ Write
            /\ feed' = feed \cup Events(attached, "log.withdrawn", now["delete"])
            /\ Finish("delete")
            /\ UNCHANGED <<row, frozen, attached, eligible, described, user, registered>>
       ELSE Abort("delete")
    /\ TimeUnchanged

Delete ==
    FindLog("delete", "lock") \/ LockUser("delete", "lock", "now") \/ Now("delete", "now", "commit") \/ DeleteCommit

\* UpdateLogContestRegistrations: FindLog before the transaction, then LockUser,
\* FindLog again and SelectRegistrationsForScoring, which picks the requested
\* registrations in contests unfinished at its own reading of the clock. Then it
\* reads the time. PlanContestRegistrationUpdate attaches the picked
\* registrations that are not attached in a contest unfinished at that time,
\* detaches the attached ones there that were not picked, and stops when there
\* is nothing to do. UpdateContestRegistrations: lockMutableLog,
\* DetachContestLog and CreateContestLog, then recomputeOfficialEligibility,
\* which runs SetLogOfficialEligibility. Picking a contest that finished between
\* the two readings while the log is still attached there inserts a duplicate
\* contest_logs row, and the transaction rolls back.
AttachSelect ==
    /\ pc["attach"] = "select"
    /\ \E requested \in SUBSET (registered \cap open) : picked' = [picked EXCEPT !["attach"] = requested]
    /\ Goto("attach", "now")
    /\ UNCHANGED <<Data, open, clock, lockU, now, seen>>

AttachCommit ==
    /\ pc["attach"] = "commit"
    /\ LET current == attached \cap seen["attach"]
           add == picked["attach"] \ current
           drop == current \ picked["attach"]
           after == (attached \ drop) \cup add
       IN IF Mutable /\ (add # {} \/ drop # {}) /\ add \cap attached = {}
          THEN /\ attached' = after
               /\ eligible' = (after \cap Official # {})
               /\ Write
               /\ feed' = feed \cup Diff(attached, after, FALSE, now["attach"])
               /\ Finish("attach")
               /\ UNCHANGED <<row, deleted, frozen, described, user, registered>>
          ELSE Abort("attach")
    /\ TimeUnchanged

Attach ==
    \/ FindLog("attach", "lock")
    \/ LockUser("attach", "lock", "select")
    \/ AttachSelect
    \/ Now("attach", "now", "commit")
    \/ AttachCommit

\* DetachContestLog: FindLog before the transaction, then LockUser, the audit
\* event and the time. ModerateDetach takes no log lock and checks neither
\* deleted_at nor frozen_at: DetachContestLog, recomputeOfficialEligibility, and
\* the diff of the views. It may detach any contest, finished or not.
ModerateCommit ==
    /\ pc["moderate"] = "commit"
    /\ \E c \in Contests :
        LET after == IF Visible THEN attached \ {c} ELSE {}
        IN /\ attached' = attached \ {c}
           /\ eligible' = (attached' \cap Official # {})
           /\ feed' = feed \cup Diff(Views, after, FALSE, now["moderate"])
    /\ Write
    /\ Finish("moderate")
    /\ UNCHANGED <<row, deleted, frozen, described, user, registered>>
    /\ TimeUnchanged

Moderate ==
    \/ FindLog("moderate", "lock")
    \/ LockUser("moderate", "lock", "now")
    \/ Now("moderate", "now", "commit")
    \/ ModerateCommit

\* UpsertContestRegistration reads the time after LockUser. ApplyRegistration
\* refuses a finished contest; dropping the log's language runs
\* DetachContestLogsForLanguages, which skips soft-deleted logs, publishes a
\* withdrawn event per detached log, and leaves eligible_official_leaderboard as it was.
RegisterCommit ==
    /\ pc["register"] = "commit"
    /\ \E c \in registered :
        IF c \in seen["register"]
        THEN LET detached == IF Visible THEN attached \cap {c} ELSE {}
             IN /\ registered' = registered \ {c}
                /\ attached' = attached \ detached
                /\ IF detached # {} THEN Write ELSE UNCHANGED wroteFrozen
                /\ feed' = feed \cup Events(detached, "log.withdrawn", now["register"])
                /\ Finish("register")
                /\ UNCHANGED <<row, deleted, frozen, eligible, described, user>>
        ELSE Abort("register")
    /\ TimeUnchanged

Register == LockUser("register", "start", "now") \/ Now("register", "now", "commit") \/ RegisterCommit

\* Filing the deletion: LockForDeletion and MarkDeletionLocked commit together
\* under the users row lock. It publishes nothing.
FileCommit ==
    /\ pc["file"] = "start"
    /\ lockU = None
    /\ user' = IF user = "active" THEN "locked" ELSE user
    /\ Goto("file", "done")
    /\ UNCHANGED <<row, deleted, frozen, attached, eligible, wroteFrozen, described, registered, feed, open,
                   clock, lockU, now, seen, picked>>

\* The delete job exists once filing commits, and reads the time before its
\* transaction. Tombstone locks the users row; a deleted user is a no-op.
DeleterNow ==
    /\ user # "active"
    /\ Now("deleter", "start", "lock")

DeleterLock ==
    /\ pc["deleter"] = "lock"
    /\ lockU = None
    /\ IF user = "locked"
       THEN lockU' = "deleter" /\ Goto("deleter", "commit")
       ELSE Goto("deleter", "done") /\ UNCHANGED lockU
    /\ UNCHANGED <<Data, open, clock, now, seen, picked>>

\* RemoveUser, in its statement order: DeleteUserContestLogsOutsideFinishedContests
\* removes the rows of a soft-deleted log and of contests unfinished at the job's
\* time; FreezeUserContestHistory, with no deleted_at guard, freezes a log that
\* still has a row and clears its description; DeleteUserLogTags;
\* DeleteUserLogsOutsideContests hard-deletes a log with none. The diff of the
\* views gives a withdrawn event per removed contest, and an update for each kept
\* one when there was a description or a tag to clear.
RemoveUser ==
    /\ pc["deleter"] = "commit"
    /\ LET before == Views
           kept == IF deleted THEN {} ELSE attached \ seen["deleter"]
           freeze == row = "present" /\ kept # {}
           purge == row = "present" /\ kept = {}
           after == IF freeze THEN kept ELSE {}
       IN /\ attached' = kept
          /\ frozen' = (frozen \/ freeze)
          /\ row' = IF purge THEN "purged" ELSE row
          /\ Write
          /\ feed' = feed \cup Diff(before, after, described, now["deleter"])
    /\ described' = FALSE
    /\ user' = "deleted"
    /\ Finish("deleter")
    /\ UNCHANGED <<deleted, eligible, registered>>
    /\ TimeUnchanged

Deleter == DeleterNow \/ DeleterLock \/ RemoveUser

\* A contest's last day ends.
Close(c) ==
    /\ c \in open
    /\ open' = open \ {c}
    /\ UNCHANGED <<Data, clock, lockU, pc, now, seen, picked>>

\* Every request and job has finished.
Done ==
    /\ \A p \in Procs : pc[p] = "done"
    /\ UNCHANGED vars

Next ==
    \/ Create \/ Update \/ Delete \/ Attach \/ Moderate \/ Register \/ FileCommit \/ Deleter
    \/ \E c \in Contests : Close(c)
    \/ Done

Spec == Init /\ [][Next]_vars

\* TLC's view of a state: a finished process never reads its clock reading again.
Live(f, idle) == [p \in Procs |-> IF pc[p] = "done" THEN idle ELSE f[p]]
View == <<row, deleted, frozen, attached, eligible, wroteFrozen, described, user, registered, open, clock,
          feed, lockU, pc, Live(now, 0), Live(seen, {}), Live(picked, {})>>

----
\* Safety

\* No request or job is in the middle of a transaction.
Quiet == \A p \in Procs : pc[p] \in {"start", "done"}

\* The feed's events for a log in contest c, in occurrence order: the latest before time t.
Previous(c, t) == {e \in feed : e.c = c /\ e.t < t /\ ~\E f \in feed : f.c = c /\ e.t < f.t /\ f.t < t}

\* Each log's events in a contest, read oldest first, start with a submission;
\* only a new submission follows a withdrawal, and a submission follows nothing else.
ValidSequence ==
    \A e \in feed :
        IF Previous(e.c, e.t) = {}
        THEN e.ev = "log.submitted"
        ELSE \A p \in Previous(e.c, e.t) :
            IF p.ev = "log.withdrawn" THEN e.ev = "log.submitted" ELSE e.ev # "log.submitted"

\* A consumer that replays the whole feed and keeps each log's latest event
\* agrees with the contest's logs once no transaction is running.
FeedPresent(c) ==
    \E e \in feed : e.c = c /\ e.ev # "log.withdrawn" /\ ~\E f \in feed : f.c = c /\ f.t > e.t

FeedAgreesWhenQuiet == Quiet => \A c \in Contests : FeedPresent(c) = (c \in Views)

\* A frozen log and its contest rows never change again.
FrozenImmutable == ~wroteFrozen

\* After deletion only the user's history in finished contests remains, frozen.
DeletedUserKeepsOnlyContestHistory ==
    user = "deleted" /\ row = "present" => frozen /\ ~deleted /\ attached # {} /\ attached \cap open = {}

\* A frozen log belongs to a deleted user: RemoveUser freezes and tombstones in
\* one transaction. LockUser refusing that user is what keeps every other writer
\* away from frozen history.
FrozenOnlyAfterDeletion == frozen => user = "deleted"

\* The official leaderboard counts a live log of a live user exactly when the
\* log is attached to an official contest.
EligibilityMatches ==
    Visible /\ user # "deleted" => eligible = (attached \cap Official # {})
====

---- MODULE AccountDeletion ----
\* Filing and running an account deletion for one user, against the requests
\* that compete for the same accounts and users rows.
\*
\* Sources:
\*   services/tadoku-api/app/api/account_deletion.go
\*   services/tadoku-api/app/worker/account_deletion.go
\*   services/tadoku-api/app/api/authz.go
\*   services/tadoku-api/features/profile/profile_service.go
\*   services/tadoku-api/features/identity/identity_service.go
\*   services/tadoku-api/sql/identity/identity.sql
\*   services/tadoku-api/sql/profile/profile.sql
\*   docs/docs/architecture/authorization.md (granting an administrator)
\*
\* Each step is one SQL statement, named after its query or function. A
\* transaction's writes become visible when it commits, and an abort releases
\* its row locks. A statement that locks an accounts row with "for update of a"
\* judges the users row from the snapshot taken when the statement started,
\* before it waited for the lock; PostgreSQL rechecks only rows it locks. The
\* delete job runs at least once: it retries, can end failed, and an operator
\* replays it. MaxFailures bounds its transient failures.
EXTENDS Naturals

CONSTANTS MaxAttempts, MaxFailures

None == "none"
Filers == {"self", "admin"}
Procs == Filers \cup {"login", "settings", "role", "grant", "worker"}

VARIABLES
    locked,      \* users.deletion_locked_at is set
    deleted,     \* users.deleted_at is set
    account,     \* the accounts row exists
    session,     \* a sessions row exists
    admin,       \* the admin relationship exists
    banned,      \* the banned relationship exists
    job,         \* the account.delete.v1 job: none, pending, running, done or failed
    attempts,    \* attempts of the current job row
    failures,    \* transient delete job failures so far
    lockA,       \* who holds the accounts row lock
    lockU,       \* who holds the users row lock
    pc,          \* each process's next step
    snap,        \* what a waiting statement's snapshot said about the users row
    sessionAfterLock, \* a session row was inserted while the user was locked
    writeAfterLock    \* a settings write committed while the user was locked

vars == <<locked, deleted, account, session, admin, banned, job, attempts, failures,
          lockA, lockU, pc, snap, sessionAfterLock, writeAfterLock>>

Usable == account /\ ~locked /\ ~deleted

TypeOK ==
    /\ locked \in BOOLEAN /\ deleted \in BOOLEAN /\ account \in BOOLEAN /\ session \in BOOLEAN
    /\ admin \in BOOLEAN /\ banned \in BOOLEAN
    /\ job \in {"none", "pending", "running", "done", "failed"}
    /\ attempts \in 0..MaxAttempts /\ failures \in 0..MaxFailures
    /\ lockA \in Procs \cup {None} /\ lockU \in Procs \cup {None}
    /\ snap \in [Procs -> BOOLEAN]

Init ==
    /\ locked = FALSE /\ deleted = FALSE /\ account = TRUE /\ session = FALSE
    /\ admin \in BOOLEAN /\ banned \in BOOLEAN /\ ~(admin /\ banned)
    /\ job = "none" /\ attempts = 0 /\ failures = 0
    /\ lockA = None /\ lockU = None
    /\ pc = [p \in Procs |-> IF p = "worker" THEN "idle" ELSE "start"]
    /\ snap = [p \in Procs |-> FALSE]
    /\ sessionAfterLock = FALSE /\ writeAfterLock = FALSE

Goto(p, step) == pc' = [pc EXCEPT ![p] = step]

\* End a transaction: release its locks. Its buffered writes are applied by the caller.
Finish(p, step) ==
    /\ Goto(p, step)
    /\ lockA' = IF lockA = p THEN None ELSE lockA
    /\ lockU' = IF lockU = p THEN None ELSE lockU

\* Take a statement snapshot of whether the users row passes the account guard.
Snapshot(p, next) ==
    /\ pc[p] = "start"
    /\ snap' = [snap EXCEPT ![p] = ~locked /\ ~deleted]
    /\ Goto(p, next)
    /\ UNCHANGED <<locked, deleted, account, session, admin, banned, job, attempts, failures,
                   lockA, lockU, sessionAfterLock, writeAfterLock>>

Unchanged == UNCHANGED <<locked, deleted, account, session, admin, banned, job, attempts, failures,
                         snap, sessionAfterLock, writeAfterLock>>

----
\* fileAccountDeletion, for a self-service or an admin request.

\* LockAccountForRequest locks the accounts row when its snapshot shows a deletable user.
FileLockAccount(p) ==
    /\ pc[p] = "lockAccount"
    /\ IF snap[p] /\ account
       THEN lockA = None /\ lockA' = p /\ Goto(p, "deletionState") /\ UNCHANGED lockU
       ELSE Goto(p, "deletionState") /\ UNCHANGED <<lockA, lockU>>
    /\ Unchanged

\* LockForDeletion reads the committed state, then locks the users row and reads it again.
FileDeletionState(p) ==
    /\ pc[p] = "deletionState"
    /\ IF locked \/ deleted THEN Finish(p, "done") ELSE Goto(p, "lockUser") /\ UNCHANGED <<lockA, lockU>>
    /\ Unchanged

FileLockUser(p) ==
    /\ pc[p] = "lockUser"
    /\ lockU = None
    /\ IF locked \/ deleted
       THEN Finish(p, "done")
       ELSE lockU' = p /\ Goto(p, "roles") /\ UNCHANGED lockA
    /\ Unchanged

\* UserRoles: an admin can't be deleted, and a banned user can't delete themselves.
\* MarkDeletionLocked, the audit event and the enqueue commit together.
FileCommit(p) ==
    /\ pc[p] = "roles"
    /\ IF admin \/ (banned /\ p = "self")
       THEN Finish(p, "done") /\ UNCHANGED <<locked, job, attempts>>
       ELSE /\ Finish(p, "done")
            /\ locked' = TRUE
            /\ job' = "pending"
            /\ attempts' = 0
    /\ UNCHANGED <<deleted, account, session, admin, banned, failures, snap, sessionAfterLock, writeAfterLock>>

File(p) == Snapshot(p, "lockAccount") \/ FileLockAccount(p) \/ FileDeletionState(p) \/ FileLockUser(p) \/ FileCommit(p)

----
\* Login: FindCredentials, then CreateSession, one statement that locks the accounts row.

LoginCredentials ==
    /\ pc["login"] = "start"
    /\ IF Usable THEN Goto("login", "createSession") ELSE Goto("login", "done")
    /\ UNCHANGED <<locked, deleted, account, session, admin, banned, job, attempts, failures,
                   lockA, lockU, snap, sessionAfterLock, writeAfterLock>>

LoginSnapshot ==
    /\ pc["login"] = "createSession"
    /\ snap' = [snap EXCEPT !["login"] = ~locked /\ ~deleted]
    /\ Goto("login", "insertSession")
    /\ UNCHANGED <<locked, deleted, account, session, admin, banned, job, attempts, failures,
                   lockA, lockU, sessionAfterLock, writeAfterLock>>

LoginInsert ==
    /\ pc["login"] = "insertSession"
    /\ lockA = None
    /\ IF account /\ snap["login"]
       THEN session' = TRUE /\ sessionAfterLock' = (sessionAfterLock \/ locked)
       ELSE UNCHANGED <<session, sessionAfterLock>>
    /\ Goto("login", "done")
    /\ UNCHANGED <<locked, deleted, account, admin, banned, job, attempts, failures,
                   lockA, lockU, snap, writeAfterLock>>

Login == LoginCredentials \/ LoginSnapshot \/ LoginInsert

----
\* A settings write: LockAccount, then the profile write's LockUser, then commit.

SettingsLockAccount ==
    /\ pc["settings"] = "lockAccount"
    /\ IF snap["settings"] /\ account
       THEN lockA = None /\ lockA' = "settings" /\ Goto("settings", "lockUser")
       ELSE Goto("settings", "done") /\ UNCHANGED lockA
    /\ UNCHANGED lockU
    /\ Unchanged

SettingsLockUser ==
    /\ pc["settings"] = "lockUser"
    /\ lockU = None
    /\ IF locked \/ deleted
       THEN Finish("settings", "done")
       ELSE lockU' = "settings" /\ Goto("settings", "commit") /\ UNCHANGED lockA
    /\ Unchanged

SettingsCommit ==
    /\ pc["settings"] = "commit"
    /\ Finish("settings", "done")
    /\ writeAfterLock' = (writeAfterLock \/ locked)
    /\ UNCHANGED <<locked, deleted, account, session, admin, banned, job, attempts, failures, snap, sessionAfterLock>>

Settings == Snapshot("settings", "lockAccount") \/ SettingsLockAccount \/ SettingsLockUser \/ SettingsCommit

----
\* UpdateRole bans the user: LockUser, refuse an admin, then SetBanned.

RoleLockUser ==
    /\ pc["role"] = "start"
    /\ lockU = None
    /\ IF locked \/ deleted
       THEN Goto("role", "done") /\ UNCHANGED lockU
       ELSE lockU' = "role" /\ Goto("role", "commit")
    /\ UNCHANGED lockA
    /\ Unchanged

RoleCommit ==
    /\ pc["role"] = "commit"
    /\ Finish("role", "done")
    /\ banned' = (banned \/ ~admin)
    /\ UNCHANGED <<locked, deleted, account, session, admin, job, attempts, failures, snap, sessionAfterLock, writeAfterLock>>

Role == RoleLockUser \/ RoleCommit

\* Operator SQL grants an administrator after locking the users row.
GrantLockUser ==
    /\ pc["grant"] = "start"
    /\ lockU = None
    /\ IF locked \/ deleted
       THEN Goto("grant", "done") /\ UNCHANGED lockU
       ELSE lockU' = "grant" /\ Goto("grant", "commit")
    /\ UNCHANGED lockA
    /\ Unchanged

GrantCommit ==
    /\ pc["grant"] = "commit"
    /\ Finish("grant", "done")
    /\ admin' = TRUE
    /\ UNCHANGED <<locked, deleted, account, session, banned, job, attempts, failures, snap, sessionAfterLock, writeAfterLock>>

Grant == GrantLockUser \/ GrantCommit

----
\* The worker's DeleteAccount: LockAccountForDeletion, Tombstone, removeAccountData, commit.

WorkerClaim ==
    /\ pc["worker"] = "idle"
    /\ job = "pending"
    /\ job' = "running"
    /\ attempts' = attempts + 1
    /\ Goto("worker", "lockAccount")
    /\ UNCHANGED <<locked, deleted, account, session, admin, banned, failures, lockA, lockU, snap, sessionAfterLock, writeAfterLock>>

WorkerLockAccount ==
    /\ pc["worker"] = "lockAccount"
    /\ IF account
       THEN lockA = None /\ lockA' = "worker"
       ELSE UNCHANGED lockA
    /\ Goto("worker", "lockUser")
    /\ UNCHANGED lockU
    /\ Unchanged

\* Tombstone: a deleted user is a no-op success, and an unlocked user is rejected.
WorkerLockUser ==
    /\ pc["worker"] = "lockUser"
    /\ lockU = None
    /\ CASE deleted -> Finish("worker", "idle") /\ job' = "done"
         [] ~locked -> Finish("worker", "idle") /\ job' = "failed"
         [] OTHER -> lockU' = "worker" /\ Goto("worker", "commit") /\ UNCHANGED <<lockA, job>>
    /\ UNCHANGED <<locked, deleted, account, session, admin, banned, attempts, failures, snap, sessionAfterLock, writeAfterLock>>

WorkerCommit ==
    /\ pc["worker"] = "commit"
    /\ Finish("worker", "idle")
    /\ deleted' = TRUE
    /\ account' = FALSE
    /\ session' = FALSE
    /\ admin' = FALSE
    /\ banned' = FALSE
    /\ job' = "done"
    /\ UNCHANGED <<locked, attempts, failures, snap, sessionAfterLock, writeAfterLock>>

\* A transient failure rolls the attempt back: retry, or fail once attempts run out.
WorkerFails ==
    /\ pc["worker"] \in {"lockAccount", "lockUser", "commit"}
    /\ failures < MaxFailures
    /\ failures' = failures + 1
    /\ job' = IF attempts >= MaxAttempts THEN "failed" ELSE "pending"
    /\ Finish("worker", "idle")
    /\ UNCHANGED <<locked, deleted, account, session, admin, banned, attempts, snap, sessionAfterLock, writeAfterLock>>

Worker == WorkerClaim \/ WorkerLockAccount \/ WorkerLockUser \/ WorkerCommit

\* An admin fixes the cause and replays the failed job as a new job row.
OperatorReplay ==
    /\ job = "failed"
    /\ locked /\ ~deleted
    /\ job' = "pending"
    /\ attempts' = 0
    /\ UNCHANGED <<locked, deleted, account, session, admin, banned, failures, lockA, lockU, pc, snap,
                   sessionAfterLock, writeAfterLock>>

\* Every request has finished and the job is not waiting for anyone.
Done ==
    /\ \A p \in Procs \ {"worker"} : pc[p] = "done"
    /\ pc["worker"] = "idle"
    /\ job \notin {"pending", "running"}
    /\ ~(job = "failed" /\ locked /\ ~deleted)
    /\ UNCHANGED vars

Next ==
    \/ \E p \in Filers : File(p)
    \/ Login \/ Settings \/ Role \/ Grant
    \/ Worker \/ WorkerFails \/ OperatorReplay
    \/ Done

Fairness ==
    /\ \A p \in Filers : WF_vars(File(p))
    /\ WF_vars(Login) /\ WF_vars(Settings) /\ WF_vars(Role) /\ WF_vars(Grant)
    /\ WF_vars(Worker)
    /\ WF_vars(OperatorReplay)

Spec == Init /\ [][Next]_vars /\ Fairness

----
\* Safety

TombstoneRequiresLock == deleted => locked

\* Filing refuses an admin, and granting refuses a locked user, both under the users row lock.
LockedNeverAdmin == locked => ~admin

DeletedHasNoAccount == deleted => ~account /\ ~session

\* A locked user always has a delete job an operator can see.
LockedHasVisibleJob == (locked /\ ~deleted) => job \in {"pending", "running", "failed"}

\* No settings write commits after the deletion lock.
NoWriteAfterLock == ~writeAfterLock

\* Liveness: a locked user is eventually deleted, given that an admin replays a failed job.
DeletionCompletes == locked ~> deleted

\* Known gaps, checked by manual configurations.

\* CreateSession can insert a session after the lock commits, because it judges the
\* users row from its snapshot. Every later request treats that session as a guest.
NoSessionAfterLock == ~sessionAfterLock

\* Granting an administrator does not check for a ban.
RolesExclusive == ~(admin /\ banned)
====

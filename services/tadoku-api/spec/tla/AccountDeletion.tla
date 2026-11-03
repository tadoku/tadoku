---- MODULE AccountDeletion ----
\* Filing and running an account deletion for one user, against the requests
\* that compete for the same accounts and users rows.
\*
\* Sources:
\*   services/tadoku-api/app/api/account_deletion.go
\*   services/tadoku-api/app/api/identity_account_deletion.go
\*   services/tadoku-api/app/api/authz.go
\*   services/tadoku-api/app/worker/account_deletion.go
\*   services/tadoku-api/app/worker/runtime.go
\*   services/tadoku-api/features/profile/profile_service.go
\*   services/tadoku-api/features/identity/identity_service.go
\*   services/tadoku-api/features/identity/identity_settings_service.go
\*   services/tadoku-api/features/identity/identity_deletion_service.go
\*   services/tadoku-api/features/authz/authz_service.go
\*   services/tadoku-api/features/authz/roles.go
\*   services/tadoku-api/sql/identity/identity.sql
\*   services/tadoku-api/sql/profile/profile.sql
\*   docs/docs/architecture/authorization.md (granting an administrator)
\*
\* Each step is one SQL statement, named after its query or function, except
\* that statements run while a transaction holds both rows' locks are merged
\* into one step. A transaction's writes become visible when it commits, and an
\* abort releases its row locks. A statement that locks an accounts row with
\* "for update of a" judges the users row from the snapshot taken when the
\* statement started, before it waited for the lock; PostgreSQL rechecks only
\* rows it locks. The delete job runs at least once: its commit and its job
\* acknowledgement are separate, it retries, can end failed, and an operator
\* replays it. MaxFailures bounds its transient failures. Password recovery
\* stores sessions the way login does and is not modelled separately.
EXTENDS Naturals

CONSTANTS MaxAttempts, MaxFailures

None == "none"
Filers == {"self", "admin"}
Procs == Filers \cup {"login", "request", "settings", "role", "grant", "worker"}

VARIABLES
    locked,      \* users.deletion_locked_at is set
    deleted,     \* users.deleted_at is set
    account,     \* the accounts row exists
    session,     \* a sessions row exists; the user starts signed in
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
    signInAfterLock,  \* a request authenticated while the user was locked
    writeAfterLock    \* a user write committed while the user was locked

vars == <<locked, deleted, account, session, admin, banned, job, attempts, failures,
          lockA, lockU, pc, snap, sessionAfterLock, signInAfterLock, writeAfterLock>>

Usable == account /\ ~locked /\ ~deleted

TypeOK ==
    /\ locked \in BOOLEAN /\ deleted \in BOOLEAN /\ account \in BOOLEAN /\ session \in BOOLEAN
    /\ admin \in BOOLEAN /\ banned \in BOOLEAN
    /\ job \in {"none", "pending", "running", "done", "failed"}
    /\ attempts \in 0..MaxAttempts /\ failures \in 0..MaxFailures
    /\ lockA \in Procs \cup {None} /\ lockU \in Procs \cup {None}
    /\ snap \in [Procs -> BOOLEAN]

Init ==
    /\ locked = FALSE /\ deleted = FALSE /\ account = TRUE /\ session = TRUE
    /\ admin \in BOOLEAN /\ banned \in BOOLEAN /\ ~(admin /\ banned)
    /\ job = "none" /\ attempts = 0 /\ failures = 0
    /\ lockA = None /\ lockU = None
    /\ pc = [p \in Procs |-> IF p = "worker" THEN "idle" ELSE "start"]
    /\ snap = [p \in Procs |-> FALSE]
    /\ sessionAfterLock = FALSE /\ signInAfterLock = FALSE /\ writeAfterLock = FALSE

Goto(p, step) == pc' = [pc EXCEPT ![p] = step]

\* End a transaction: release its locks. Its buffered writes are applied by the caller.
Finish(p, step) ==
    /\ Goto(p, step)
    /\ lockA' = IF lockA = p THEN None ELSE lockA
    /\ lockU' = IF lockU = p THEN None ELSE lockU

\* Take a statement snapshot of whether the users row passes the account guard.
Snapshot(p, from, next) ==
    /\ pc[p] = from
    /\ snap' = [snap EXCEPT ![p] = ~locked /\ ~deleted]
    /\ Goto(p, next)
    /\ UNCHANGED <<locked, deleted, account, session, admin, banned, job, attempts, failures,
                   lockA, lockU, sessionAfterLock, signInAfterLock, writeAfterLock>>

Unchanged == UNCHANGED <<locked, deleted, account, session, admin, banned, job, attempts, failures,
                         snap, sessionAfterLock, signInAfterLock, writeAfterLock>>

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
    /\ UNCHANGED <<deleted, account, session, admin, banned, failures, snap, sessionAfterLock, signInAfterLock, writeAfterLock>>

File(p) == Snapshot(p, "start", "lockAccount") \/ FileLockAccount(p) \/ FileDeletionState(p) \/ FileLockUser(p) \/ FileCommit(p)

----
\* Login: FindCredentials, then CreateSession, one statement that locks the
\* accounts row. When CreateSession inserts a row, login reports success.

LoginCredentials ==
    /\ pc["login"] = "start"
    /\ IF Usable THEN Goto("login", "createSession") ELSE Goto("login", "done")
    /\ UNCHANGED <<locked, deleted, account, session, admin, banned, job, attempts, failures,
                   lockA, lockU, snap, sessionAfterLock, signInAfterLock, writeAfterLock>>

LoginInsert ==
    /\ pc["login"] = "insertSession"
    /\ lockA = None
    /\ IF account /\ snap["login"]
       THEN session' = TRUE /\ sessionAfterLock' = (sessionAfterLock \/ locked)
       ELSE UNCHANGED <<session, sessionAfterLock>>
    /\ Goto("login", "done")
    /\ UNCHANGED <<locked, deleted, account, admin, banned, job, attempts, failures,
                   lockA, lockU, snap, signInAfterLock, writeAfterLock>>

Login == LoginCredentials \/ Snapshot("login", "createSession", "insertSession") \/ LoginInsert

\* A later request authenticates with the session: FindSession reads the latest
\* committed state, and a session of an account that is not usable is a guest.
RequestFindSession ==
    /\ pc["request"] = "start"
    /\ signInAfterLock' = (signInAfterLock \/ (session /\ Usable /\ locked))
    /\ Goto("request", "done")
    /\ UNCHANGED <<locked, deleted, account, session, admin, banned, job, attempts, failures,
                   lockA, lockU, snap, sessionAfterLock, writeAfterLock>>

----
\* A settings write (lockSettingsSession, then the write): FindSession,
\* LockAccount, FindSession again, then UpdateDisplayName on the users row.

SettingsSession ==
    /\ pc["settings"] = "start"
    /\ IF session /\ Usable THEN Goto("settings", "snapshot") ELSE Goto("settings", "done")
    /\ UNCHANGED <<lockA, lockU>>
    /\ Unchanged

SettingsLockAccount ==
    /\ pc["settings"] = "lockAccount"
    /\ IF snap["settings"] /\ account
       THEN lockA = None /\ lockA' = "settings" /\ Goto("settings", "recheck")
       ELSE Goto("settings", "done") /\ UNCHANGED lockA
    /\ UNCHANGED lockU
    /\ Unchanged

\* The session is looked up again after the account lock, from a new snapshot.
SettingsRecheck ==
    /\ pc["settings"] = "recheck"
    /\ IF session /\ Usable
       THEN Goto("settings", "update") /\ UNCHANGED <<lockA, lockU>>
       ELSE Finish("settings", "done")
    /\ Unchanged

SettingsUpdate ==
    /\ pc["settings"] = "update"
    /\ lockU = None
    /\ lockU' = "settings"
    /\ Goto("settings", "commit")
    /\ UNCHANGED lockA
    /\ Unchanged

SettingsCommit ==
    /\ pc["settings"] = "commit"
    /\ Finish("settings", "done")
    /\ writeAfterLock' = (writeAfterLock \/ locked)
    /\ UNCHANGED <<locked, deleted, account, session, admin, banned, job, attempts, failures, snap,
                   sessionAfterLock, signInAfterLock>>

Settings ==
    \/ SettingsSession
    \/ Snapshot("settings", "snapshot", "lockAccount")
    \/ SettingsLockAccount
    \/ SettingsRecheck
    \/ SettingsUpdate
    \/ SettingsCommit

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
    /\ writeAfterLock' = (writeAfterLock \/ locked)
    /\ UNCHANGED <<locked, deleted, account, session, admin, job, attempts, failures, snap,
                   sessionAfterLock, signInAfterLock>>

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
    /\ writeAfterLock' = (writeAfterLock \/ locked)
    /\ UNCHANGED <<locked, deleted, account, session, banned, job, attempts, failures, snap,
                   sessionAfterLock, signInAfterLock>>

Grant == GrantLockUser \/ GrantCommit

----
\* The worker's DeleteAccount: LockAccountForDeletion, Tombstone, removeAccountData
\* and commit, then the job runtime's separate Complete.

WorkerClaim ==
    /\ pc["worker"] = "idle"
    /\ job = "pending"
    /\ job' = "running"
    /\ attempts' = attempts + 1
    /\ Goto("worker", "lockAccount")
    /\ UNCHANGED <<locked, deleted, account, session, admin, banned, failures, lockA, lockU, snap,
                   sessionAfterLock, signInAfterLock, writeAfterLock>>

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
    /\ CASE deleted -> Finish("worker", "ack") /\ UNCHANGED job
         [] ~locked -> Finish("worker", "idle") /\ job' = "failed"
         [] OTHER -> lockU' = "worker" /\ Goto("worker", "commit") /\ UNCHANGED <<lockA, job>>
    /\ UNCHANGED <<locked, deleted, account, session, admin, banned, attempts, failures, snap,
                   sessionAfterLock, signInAfterLock, writeAfterLock>>

WorkerCommit ==
    /\ pc["worker"] = "commit"
    /\ Finish("worker", "ack")
    /\ deleted' = TRUE
    /\ account' = FALSE
    /\ session' = FALSE
    /\ admin' = FALSE
    /\ banned' = FALSE
    /\ UNCHANGED <<locked, job, attempts, failures, snap, sessionAfterLock, signInAfterLock, writeAfterLock>>

\* The job runtime records success in its own transaction.
WorkerAck ==
    /\ pc["worker"] = "ack"
    /\ job' = "done"
    /\ Goto("worker", "idle")
    /\ UNCHANGED <<locked, deleted, account, session, admin, banned, attempts, failures, lockA, lockU, snap,
                   sessionAfterLock, signInAfterLock, writeAfterLock>>

\* A transient failure: before the commit it rolls the attempt back, and after
\* it the acknowledgement is lost. Either way the job retries, or fails once
\* attempts run out.
WorkerFails ==
    /\ pc["worker"] \in {"lockAccount", "lockUser", "commit", "ack"}
    /\ failures < MaxFailures
    /\ failures' = failures + 1
    /\ job' = IF attempts >= MaxAttempts THEN "failed" ELSE "pending"
    /\ Finish("worker", "idle")
    /\ UNCHANGED <<locked, deleted, account, session, admin, banned, attempts, snap,
                   sessionAfterLock, signInAfterLock, writeAfterLock>>

Worker == WorkerClaim \/ WorkerLockAccount \/ WorkerLockUser \/ WorkerCommit \/ WorkerAck

\* An admin fixes the cause and replays the failed job as a new job row, even
\* when the deletion had already committed before the failure.
OperatorReplay ==
    /\ job = "failed"
    /\ job' = "pending"
    /\ attempts' = 0
    /\ UNCHANGED <<locked, deleted, account, session, admin, banned, failures, lockA, lockU, pc, snap,
                   sessionAfterLock, signInAfterLock, writeAfterLock>>

\* Every request has finished and no delete job is waiting.
Done ==
    /\ \A p \in Procs \ {"worker"} : pc[p] = "done"
    /\ pc["worker"] = "idle"
    /\ job \in {"none", "done"}
    /\ UNCHANGED vars

Next ==
    \/ \E p \in Filers : File(p)
    \/ Login \/ RequestFindSession \/ Settings \/ Role \/ Grant
    \/ Worker \/ WorkerFails \/ OperatorReplay
    \/ Done

Fairness ==
    /\ \A p \in Filers : WF_vars(File(p))
    /\ WF_vars(Login) /\ WF_vars(RequestFindSession) /\ WF_vars(Settings) /\ WF_vars(Role) /\ WF_vars(Grant)
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

\* No settings write, ban or administrator grant commits after the deletion lock.
NoWriteAfterLock == ~writeAfterLock

\* No request authenticates after the deletion lock.
NoSignInAfterLock == ~signInAfterLock

\* Liveness: a locked user is eventually deleted, given that an admin replays a failed job.
DeletionCompletes == locked ~> deleted

\* Known gaps, checked by manual configurations.

\* CreateSession can insert a session after the lock commits, because it judges the
\* users row from its snapshot, and login then reports success. NoSignInAfterLock
\* checks that every later request treats that session as a guest.
NoSessionAfterLock == ~sessionAfterLock

\* Granting an administrator does not check for a ban.
RolesExclusive == ~(admin /\ banned)
====

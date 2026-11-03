---- MODULE LeaderboardCache ----
\* The Valkey cache of unfiltered leaderboards: requests that read it or
\* rebuild it from PostgreSQL behind a generation check, against the writes
\* and invalidation jobs that make it stale.
\*
\* Sources:
\*   services/tadoku-api/features/leaderboard/leaderboard_cache.go
\*   services/tadoku-api/features/leaderboard/leaderboard_service.go
\*   services/tadoku-api/features/leaderboard/leaderboard_repository.go
\*   services/tadoku-api/sql/leaderboard/leaderboard.sql
\*   services/tadoku-api/app/api/log_mutations.go
\*   services/tadoku-api/app/api/log_attachments.go
\*   services/tadoku-api/app/api/contests.go
\*   services/tadoku-api/app/worker/application.go (InvalidateContestLeaderboard, InvalidateOfficialLeaderboard)
\*   services/tadoku-api/features/logs/logs_mutation_service.go
\*   services/tadoku-api/features/logs/logs_removal_service.go
\*   services/tadoku-api/features/contests/contests_service.go
\*   services/tadoku-api/features/contests/contests_removal_service.go
\*   services/tadoku-api/app/worker/account_deletion.go
\*   services/tadoku-api/spec/tla/JobQueue.tla (the job contract)
\*   k8s/dev/base/data/cache.yaml (allkeys-lru, no persistence)
\*
\* A board has three Valkey keys: the sorted set, its ":last_updated" marker
\* and its ":generation" counter. Each Valkey command a request sends is one
\* step, and each Lua script (rebuildScript, pageScript, invalidateScript) is
\* one atomic step that keeps the writes it made before an error.
\* rebuildScript adds scores in chunks small enough for Lua's unpack, and
\* Valkey refuses a memory-growing or replica write only before a script's
\* first write, so the script cannot fail after its first write for numeric
\* scores. A NaN score would fail ZADD after the DEL and leave a truncated set
\* beside the marker; the model leaves that out. Verified on Valkey 9;
\* production's Upstash runtime is not verified.
\*
\* Every producer enqueues its invalidation job in the transaction of the write
\* that changes a board, so a write is one step that raises the board's
\* PostgreSQL version and enqueues the job. Version 0 is the empty board, which
\* Valkey stores as an absent sorted set, so an invalidated or evicted set
\* reads as version 0. A missing generation reads as "0", and INCR starts it at
\* 1. A request's PostgreSQL read is one statement, which sees the versions
\* committed when it started; ContestExists is folded into it.
\*
\* An invalidation job invalidates the keys in Keys in order, one script each.
\* A job runs at least once: an attempt can fail before or after any script
\* applies, retries until MaxAttempts and then fails, and an operator replays a
\* failed job as a new row. MaxFailures bounds transient failures. A duplicate
\* run of one job acts on a key like a run of another job on that key, which
\* the configurations with two writes cover; JobQueue.tla covers duplicates.
\* Any Valkey error sends a request to PostgreSQL, which changes no cache
\* state. Filtered reads only read PostgreSQL and are not modelled.
\*
\* rebuildScript sets the marker with a 28-day expiry. MarkerExpires lets a
\* marker lapse at any point, and fairness makes every marker lapse eventually;
\* a later rebuild sets a new one.
\*
\* Losses lists how Valkey may lose keys, and MaxLosses bounds how often:
\* eviction of one key ("data", "marker" or "generation"), since allkeys-lru
\* evicts keys one at a time, or a "restart" without persistence that drops
\* every key. Production's eviction setting is not in this repository; CI
\* checks a Valkey that never loses keys.
EXTENDS Integers, Sequences

CONSTANTS OfficialJob, Readers, MaxWrites, MaxAttempts, MaxFailures, Losses, MaxLosses, OperatorReplays

\* The keys a job invalidates, in order: InvalidateOfficial's yearly then
\* global board, or InvalidateContest's one board.
Keys == IF OfficialJob THEN <<"yearly", "global">> ELSE <<"contest">>
KeySet == {Keys[i] : i \in 1..Len(Keys)}
Jobs == 1..MaxWrites
Versions == 0..MaxWrites
Absent == -1
Parts == {"data", "marker", "generation"}

VARIABLES
    pg,         \* the board's version in PostgreSQL: committed writes
    data,       \* the version each key's sorted set holds; 0 when absent
    marker,     \* the generation each ":last_updated" marker names, or Absent
    gen,        \* each ":generation" counter; 0 when absent
    covered,    \* the PostgreSQL version when an invalidation last applied to each key
    job,        \* each write's invalidation job: none, pending, running, done or failed
    step,       \* the position in Keys of the running job's next script
    attempts,   \* attempts of the job's current row
    failures,   \* transient invalidation failures so far
    losses,     \* lost Valkey keys so far
    req,        \* each request: its next step, key, floor, first marker, generation, page and loaded version
    staleServe  \* a request served a cached board older than its floor

vars == <<pg, data, marker, gen, covered, job, step, attempts, failures, losses,
          req, staleServe>>
cache == <<data, marker, gen, covered>>
jobVars == <<job, step, attempts, failures>>
ReqIdle == [pc |-> "idle", key |-> Keys[1], floor |-> 0, marker |-> Absent, gen |-> 0, data |-> 0, load |-> 0]

MaxGen == MaxWrites * Len(Keys) * (MaxAttempts + MaxFailures)

TypeOK ==
    /\ pg \in Versions
    /\ data \in [KeySet -> Versions]
    /\ marker \in [KeySet -> {Absent} \cup 0..MaxGen]
    /\ gen \in [KeySet -> 0..MaxGen]
    /\ covered \in [KeySet -> Versions]
    /\ job \in [Jobs -> {"none", "pending", "running", "done", "failed"}]
    /\ step \in [Jobs -> 1..(Len(Keys) + 1)]
    /\ attempts \in [Jobs -> 0..MaxAttempts]
    /\ failures \in 0..MaxFailures
    /\ losses \in 0..MaxLosses
    /\ Losses \subseteq Parts \cup {"restart"}
    /\ \A r \in Readers : req[r].pc \in {"idle", "gen", "page", "remarker", "regen", "missgen", "load", "rebuild"}
    /\ staleServe \in BOOLEAN

Init ==
    /\ pg = 0
    /\ data = [k \in KeySet |-> 0]
    /\ marker = [k \in KeySet |-> Absent]
    /\ gen = [k \in KeySet |-> 0]
    /\ covered = [k \in KeySet |-> 0]
    /\ job = [j \in Jobs |-> "none"]
    /\ step = [j \in Jobs |-> 1]
    /\ attempts = [j \in Jobs |-> 0]
    /\ failures = 0
    /\ losses = 0
    /\ req = [r \in Readers |-> ReqIdle]
    /\ staleServe = FALSE

Set(r, rec) == req' = [req EXCEPT ![r] = rec]

\* A request step that changes only the request's own state.
ReqOnly == UNCHANGED <<pg, cache, jobVars, losses, staleServe>>

\* fetchPage accepts a marker naming the current generation, or any marker when
\* the generation is "0".
Accepts(m, g) == g = 0 \/ m = g

\* The marker and generation on Valkey would let fetchPage serve the key.
Servable(k) == marker[k] # Absent /\ Accepts(marker[k], gen[k])

----
\* A write to logs, contest logs or registrations commits with its invalidation job.
Write ==
    /\ pg < MaxWrites
    /\ pg' = pg + 1
    /\ job' = [job EXCEPT ![pg + 1] = "pending"]
    /\ UNCHANGED <<cache, step, attempts, failures, losses, req, staleServe>>

----
\* Service.fetch for an unfiltered board. Each step is one Valkey command or
\* script, or the PostgreSQL read.

\* fetchPage: GET ":last_updated". A missing marker is a miss.
ReadMarker(r, k) ==
    /\ req[r].pc = "idle"
    /\ Set(r, [ReqIdle EXCEPT !.pc = IF marker[k] = Absent THEN "missgen" ELSE "gen",
                              !.key = k,
                              !.floor = covered[k],
                              !.marker = marker[k]])
    /\ ReqOnly

\* fetchPage: GET ":generation", then the marker check.
ReadGeneration(r) ==
    LET k == req[r].key IN
    /\ req[r].pc = "gen"
    /\ Set(r, [req[r] EXCEPT !.pc = IF Accepts(req[r].marker, gen[k]) THEN "page" ELSE "missgen",
                             !.gen = gen[k]])
    /\ ReqOnly

\* pageScript reads the page and its counts from one snapshot of the set.
PageScript(r) ==
    /\ req[r].pc = "page"
    /\ Set(r, [req[r] EXCEPT !.pc = "remarker", !.data = data[req[r].key]])
    /\ ReqOnly

\* cacheCurrent: GET ":last_updated" again; a changed or missing marker is a miss.
RecheckMarker(r) ==
    /\ req[r].pc = "remarker"
    /\ Set(r, [req[r] EXCEPT !.pc = IF marker[req[r].key] = req[r].marker THEN "regen" ELSE "missgen"])
    /\ ReqOnly

\* cacheCurrent: GET ":generation" again. An unchanged generation serves the page.
RecheckGeneration(r) ==
    LET k == req[r].key IN
    /\ req[r].pc = "regen"
    /\ IF gen[k] = req[r].gen
       THEN /\ Set(r, ReqIdle)
            /\ staleServe' = (staleServe \/ req[r].data < req[r].floor)
       ELSE Set(r, [req[r] EXCEPT !.pc = "missgen"]) /\ UNCHANGED staleServe
    /\ UNCHANGED <<pg, cache, jobVars, losses>>

\* On a miss, fetch reads the generation before it loads the board.
MissGeneration(r) ==
    /\ req[r].pc = "missgen"
    /\ Set(r, [req[r] EXCEPT !.pc = "load", !.gen = gen[req[r].key]])
    /\ ReqOnly

\* ContestLeaderboard or OfficialLeaderboard.
Load(r) ==
    /\ req[r].pc = "load"
    /\ Set(r, [req[r] EXCEPT !.pc = "rebuild", !.load = pg])
    /\ ReqOnly

\* rebuildScript publishes the loaded board only while the generation is the one
\* read before the load. The request serves its PostgreSQL board either way.
RebuildScript(r) ==
    LET k == req[r].key IN
    /\ req[r].pc = "rebuild"
    /\ IF gen[k] = req[r].gen
       THEN /\ data' = [data EXCEPT ![k] = req[r].load]
            /\ marker' = [marker EXCEPT ![k] = req[r].gen]
       ELSE UNCHANGED <<data, marker>>
    /\ Set(r, ReqIdle)
    /\ UNCHANGED <<pg, gen, covered, jobVars, losses, staleServe>>

\* A Valkey error or timeout: the request reads PostgreSQL instead. A rebuild
\* that errors after its script applied is RebuildScript.
ValkeyError(r) ==
    /\ req[r].pc \in {"gen", "page", "remarker", "regen", "missgen", "rebuild"}
    /\ Set(r, ReqIdle)
    /\ ReqOnly

Request(r) ==
    \/ \E k \in KeySet : ReadMarker(r, k)
    \/ ReadGeneration(r) \/ PageScript(r) \/ RecheckMarker(r) \/ RecheckGeneration(r)
    \/ MissGeneration(r) \/ Load(r) \/ RebuildScript(r) \/ ValkeyError(r)

----
\* The worker's InvalidateContestLeaderboard or InvalidateOfficialLeaderboard.

Claim(j) ==
    /\ job[j] = "pending"
    /\ job' = [job EXCEPT ![j] = "running"]
    /\ step' = [step EXCEPT ![j] = 1]
    /\ attempts' = [attempts EXCEPT ![j] = @ + 1]
    /\ UNCHANGED <<pg, cache, failures, losses, req, staleServe>>

\* invalidateScript: INCR ":generation", then DEL the set and its marker.
InvalidateScript(j) ==
    LET k == Keys[step[j]] IN
    /\ job[j] = "running"
    /\ step[j] <= Len(Keys)
    /\ gen' = [gen EXCEPT ![k] = @ + 1]
    /\ data' = [data EXCEPT ![k] = 0]
    /\ marker' = [marker EXCEPT ![k] = Absent]
    /\ covered' = [covered EXCEPT ![k] = pg]
    /\ step' = [step EXCEPT ![j] = @ + 1]
    /\ UNCHANGED <<pg, job, attempts, failures, losses, req, staleServe>>

Complete(j) ==
    /\ job[j] = "running"
    /\ step[j] = Len(Keys) + 1
    /\ job' = [job EXCEPT ![j] = "done"]
    /\ UNCHANGED <<pg, cache, step, attempts, failures, losses, req, staleServe>>

\* An attempt fails before its next script, or after its last one when the
\* acknowledgement is lost. It retries, or fails once attempts run out.
JobFails(j) ==
    /\ job[j] = "running"
    /\ failures < MaxFailures
    /\ failures' = failures + 1
    /\ job' = [job EXCEPT ![j] = IF attempts[j] >= MaxAttempts THEN "failed" ELSE "pending"]
    /\ UNCHANGED <<pg, cache, step, attempts, losses, req, staleServe>>

\* An operator replays the failed job as a new row.
OperatorReplay(j) ==
    /\ job[j] = "failed"
    /\ job' = [job EXCEPT ![j] = "pending"]
    /\ attempts' = [attempts EXCEPT ![j] = 0]
    /\ UNCHANGED <<pg, cache, step, failures, losses, req, staleServe>>

Invalidation(j) == Claim(j) \/ InvalidateScript(j) \/ Complete(j)

----
\* Valkey loses a key: allkeys-lru evicts one, or a restart without
\* persistence drops every key.

Evict(k, part) ==
    /\ part \in Losses
    /\ losses < MaxLosses
    /\ losses' = losses + 1
    /\ data' = IF part = "data" THEN [data EXCEPT ![k] = 0] ELSE data
    /\ marker' = IF part = "marker" THEN [marker EXCEPT ![k] = Absent] ELSE marker
    /\ gen' = IF part = "generation" THEN [gen EXCEPT ![k] = 0] ELSE gen
    /\ UNCHANGED <<pg, covered, jobVars, req, staleServe>>

\* The marker's 28-day expiry runs out.
MarkerExpires(k) ==
    /\ marker[k] # Absent
    /\ marker' = [marker EXCEPT ![k] = Absent]
    /\ UNCHANGED <<pg, data, gen, covered, jobVars, losses, req, staleServe>>

Restart ==
    /\ "restart" \in Losses
    /\ losses < MaxLosses
    /\ losses' = losses + 1
    /\ data' = [k \in KeySet |-> 0]
    /\ marker' = [k \in KeySet |-> Absent]
    /\ gen' = [k \in KeySet |-> 0]
    /\ UNCHANGED <<pg, covered, jobVars, req, staleServe>>

Next ==
    \/ Write
    \/ \E r \in Readers : Request(r)
    \/ \E j \in Jobs : Invalidation(j) \/ JobFails(j) \/ OperatorReplay(j)
    \/ \E k \in KeySet : MarkerExpires(k)
    \/ \E k \in KeySet, part \in Parts : Evict(k, part)
    \/ Restart

\* Jobs keep running and markers expire. With OperatorReplays, an operator
\* eventually replays a failed job. Requests, writes, failures and losses are
\* never forced.
Fairness ==
    /\ \A j \in Jobs : WF_vars(Invalidation(j))
    /\ \A k \in KeySet : WF_vars(MarkerExpires(k))
    /\ \A j \in Jobs : IF OperatorReplays THEN WF_vars(OperatorReplay(j)) ELSE TRUE

Spec == Init /\ [][Next]_vars /\ Fairness

----
\* Safety

\* A request never serves a cached board older than an invalidation that had
\* applied to that key before the request started, its floor. An invalidation
\* that applies while the request runs is concurrent with it.
ServedFenced == ~staleServe

\* A key the cache would serve never holds a board older than the last
\* invalidation applied to it.
PublishedFenced == \A k \in KeySet : Servable(k) => data[k] >= covered[k]

\* A marker always names the key's current generation: an invalidation deletes
\* it, and a rebuild writes only the generation it checked. While Valkey keeps
\* a board's keys, fetchPage's exception for generation "0" therefore accepts
\* only the marker "native:0".
MarkerNamesGeneration == \A k \in KeySet : marker[k] \in {Absent, gen[k]}

GenerationMonotonic == [][\A k \in KeySet : gen'[k] >= gen[k]]_vars

\* Liveness: once writes stop, every key the cache would serve comes to hold
\* the PostgreSQL board, through a completed invalidation or the marker's
\* expiry, even when no operator replays a failed job.
Fresh == \A k \in KeySet : Servable(k) => data[k] = pg

Converges == <>[]Fresh
====

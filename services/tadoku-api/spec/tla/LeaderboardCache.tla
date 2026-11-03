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
\* and its ":generation". Each Valkey command a request sends is one step, and
\* each Lua script (generationScript, rebuildScript, pageScript,
\* invalidateScript) is one atomic step that keeps the writes it made before an
\* error. pageScript reads all three keys at once and serves the page only when
\* the marker names the current generation and the set's size, so a lost key
\* is a miss. A generation is a random value: generationScript creates one when
\* it is missing, before a request loads, and invalidateScript increments it,
\* or sets a new random value when it is missing. Random values are assumed
\* never to collide, so a new generation differs from every value a key, marker
\* or request holds: a lost generation never comes back to a value a request
\* read. The model reuses only values none of them holds, which no comparison
\* tells from fresh ones. rebuildScript publishes only under the generation
\* the request read: it deletes the marker and the set, adds the scores, then
\* writes the marker with the set's size. It adds scores in
\* chunks small enough for Lua's unpack, and Valkey refuses a memory-growing or
\* replica write only before a script's first write; only a NaN score fails
\* after the DELs, leaving a truncated set without a marker. Verified on Valkey
\* 9; production's Upstash runtime is not verified.
\*
\* Every producer enqueues its invalidation job in the transaction of the write
\* that changes a board, so a write is one step that raises the board's
\* PostgreSQL version and enqueues the job. Version 0 is the empty board, which
\* Valkey stores as an absent sorted set, so an invalidated or evicted set
\* reads as version 0. Every other board, and a truncated set, has the same
\* size: the worst case for the size check. A request's
\* PostgreSQL read is one statement, which sees the versions committed when it
\* started; ContestExists is folded into it.
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
\* rebuildScript sets the marker with a 28-day expiry. A generation expires an
\* hour after a script creates it, unless a rebuild publishes under it, which
\* gives it the marker's expiry. With Expiry, MarkerExpires lets a marker lapse
\* at any point and GenerationExpires lets a generation without a marker lapse;
\* fairness makes every marker lapse eventually, and a later rebuild sets a new
\* one. Configurations without it check that invalidation alone makes boards
\* fresh. generationScript also replaces a value below 2^32, which only the
\* previous release's INCR produces; the model has no previous release.
\*
\* Faults lists what may go wrong in Valkey, and MaxFaults bounds how often:
\* eviction of one key ("data", "marker" or "generation"), since allkeys-lru
\* evicts keys one at a time, a "restart" without persistence that drops every
\* key, or a "truncate", a rebuild that a NaN score stops after its DEL.
\* Production's eviction setting is not in this repository.
EXTENDS Integers, Sequences, FiniteSets

CONSTANTS OfficialJob, Readers, MaxWrites, MaxAttempts, MaxFailures, Faults, MaxFaults, OperatorReplays, Expiry

\* The keys a job invalidates, in order: InvalidateOfficial's yearly then
\* global board, or InvalidateContest's one board.
Keys == IF OfficialJob THEN <<"yearly", "global">> ELSE <<"contest">>
KeySet == {Keys[i] : i \in 1..Len(Keys)}
Jobs == 1..MaxWrites
Versions == 0..MaxWrites
Truncated == -1
Parts == {"data", "marker", "generation"}
NoMarker == <<>>
NoGeneration == 0

\* A set's size as far as the marker can tell boards apart.
Size(v) == IF v = 0 THEN 0 ELSE 1

MaxGen == 2 * Len(Keys) + Cardinality(Readers) + 1

VARIABLES
    pg,         \* the board's version in PostgreSQL: committed writes
    data,       \* the version each key's sorted set holds; 0 when absent
    marker,     \* each ":last_updated" marker as <<generation, size>>, or NoMarker
    gen,        \* each ":generation", or NoGeneration
    covered,    \* the PostgreSQL version when an invalidation last applied to each key
    job,        \* each write's invalidation job: none, pending, running, done or failed
    step,       \* the position in Keys of the running job's next script
    attempts,   \* attempts of the job's current row
    failures,   \* transient invalidation failures so far
    faults,     \* Valkey faults so far
    req,        \* each request: its next step, key, generation and loaded version
    staleServe  \* a request served a cached board older than an invalidation before it

vars == <<pg, data, marker, gen, covered, job, step, attempts, failures, faults, req, staleServe>>
cache == <<data, marker, gen, covered>>
jobVars == <<job, step, attempts, failures>>
ReqIdle == [pc |-> "idle", key |-> Keys[1], gen |-> NoGeneration, load |-> 0]

TypeOK ==
    /\ pg \in Versions
    /\ data \in [KeySet -> Versions \cup {Truncated}]
    /\ marker \in [KeySet -> {NoMarker} \cup ((1..MaxGen) \X {0, 1})]
    /\ gen \in [KeySet -> {NoGeneration} \cup 1..MaxGen]
    /\ covered \in [KeySet -> Versions]
    /\ job \in [Jobs -> {"none", "pending", "running", "done", "failed"}]
    /\ step \in [Jobs -> 1..(Len(Keys) + 1)]
    /\ attempts \in [Jobs -> 0..MaxAttempts]
    /\ failures \in 0..MaxFailures
    /\ faults \in 0..MaxFaults
    /\ Faults \subseteq Parts \cup {"restart", "truncate"}
    /\ \A r \in Readers : req[r].pc \in {"idle", "missgen", "load", "rebuild"}
    /\ staleServe \in BOOLEAN

Init ==
    /\ pg = 0
    /\ data = [k \in KeySet |-> 0]
    /\ marker = [k \in KeySet |-> NoMarker]
    /\ gen = [k \in KeySet |-> NoGeneration]
    /\ covered = [k \in KeySet |-> 0]
    /\ job = [j \in Jobs |-> "none"]
    /\ step = [j \in Jobs |-> 1]
    /\ attempts = [j \in Jobs |-> 0]
    /\ failures = 0
    /\ faults = 0
    /\ req = [r \in Readers |-> ReqIdle]
    /\ staleServe = FALSE

Set(r, rec) == req' = [req EXCEPT ![r] = rec]

\* The generations a key, marker or request holds, and a value none of them
\* holds.
Held ==
    {gen[k] : k \in KeySet}
    \cup {marker[k][1] : k \in {m \in KeySet : marker[m] # NoMarker}}
    \cup {req[r].gen : r \in Readers}
NewGen == CHOOSE v \in 1..MaxGen : v \notin Held

\* A request step that changes only the request's own state.
ReqOnly == UNCHANGED <<pg, cache, jobVars, faults, staleServe>>

\* pageScript would serve the key: its marker names the current generation and
\* the set's size.
Servable(k) == gen[k] # NoGeneration /\ marker[k] = <<gen[k], Size(data[k])>>

----
\* A write to logs, contest logs or registrations commits with its invalidation job.
Write ==
    /\ pg < MaxWrites
    /\ pg' = pg + 1
    /\ job' = [job EXCEPT ![pg + 1] = "pending"]
    /\ UNCHANGED <<cache, step, attempts, failures, faults, req, staleServe>>

----
\* Service.fetch for an unfiltered board. Each step is one Valkey script or the
\* PostgreSQL read.

\* fetchPage: pageScript serves the page, or reports a miss.
PageScript(r, k) ==
    /\ req[r].pc = "idle"
    /\ IF Servable(k)
       THEN /\ Set(r, ReqIdle)
            /\ staleServe' = (staleServe \/ data[k] < covered[k])
       ELSE /\ Set(r, [ReqIdle EXCEPT !.pc = "missgen", !.key = k])
            /\ UNCHANGED staleServe
    /\ UNCHANGED <<pg, cache, jobVars, faults>>

\* On a miss, generationScript returns the generation, creating a random one
\* when it is missing, before the request loads the board.
GenerationScript(r) ==
    LET k == req[r].key IN
    /\ req[r].pc = "missgen"
    /\ IF gen[k] = NoGeneration
       THEN /\ gen' = [gen EXCEPT ![k] = NewGen]
            /\ Set(r, [req[r] EXCEPT !.pc = "load", !.gen = NewGen])
       ELSE /\ Set(r, [req[r] EXCEPT !.pc = "load", !.gen = gen[k]])
            /\ UNCHANGED gen
    /\ UNCHANGED <<pg, data, marker, covered, jobVars, faults, staleServe>>

\* ContestLeaderboard or OfficialLeaderboard.
Load(r) ==
    /\ req[r].pc = "load"
    /\ Set(r, [req[r] EXCEPT !.pc = "rebuild", !.load = pg])
    /\ ReqOnly

\* rebuildScript publishes the loaded board and its size only while the
\* generation is the one the request read. The request serves its PostgreSQL
\* board either way.
RebuildScript(r) ==
    LET k == req[r].key IN
    /\ req[r].pc = "rebuild"
    /\ IF gen[k] = req[r].gen
       THEN /\ data' = [data EXCEPT ![k] = req[r].load]
            /\ marker' = [marker EXCEPT ![k] = <<req[r].gen, Size(req[r].load)>>]
       ELSE UNCHANGED <<data, marker>>
    /\ Set(r, ReqIdle)
    /\ UNCHANGED <<pg, gen, covered, jobVars, faults, staleServe>>

\* A NaN score stops rebuildScript after its DELs and some chunks: the set is
\* truncated and the marker is gone.
RebuildScriptTruncates(r) ==
    LET k == req[r].key IN
    /\ req[r].pc = "rebuild"
    /\ "truncate" \in Faults
    /\ faults < MaxFaults
    /\ faults' = faults + 1
    /\ gen[k] = req[r].gen
    /\ data' = [data EXCEPT ![k] = Truncated]
    /\ marker' = [marker EXCEPT ![k] = NoMarker]
    /\ Set(r, ReqIdle)
    /\ UNCHANGED <<pg, gen, covered, jobVars, staleServe>>

\* A Valkey error or timeout: the request reads PostgreSQL instead. A rebuild
\* that errors after its script applied is RebuildScript.
ValkeyError(r) ==
    /\ req[r].pc \in {"missgen", "rebuild"}
    /\ Set(r, ReqIdle)
    /\ ReqOnly

Request(r) ==
    \/ \E k \in KeySet : PageScript(r, k)
    \/ GenerationScript(r) \/ Load(r) \/ RebuildScript(r) \/ RebuildScriptTruncates(r) \/ ValkeyError(r)

----
\* The worker's InvalidateContestLeaderboard or InvalidateOfficialLeaderboard.

Claim(j) ==
    /\ job[j] = "pending"
    /\ job' = [job EXCEPT ![j] = "running"]
    /\ step' = [step EXCEPT ![j] = 1]
    /\ attempts' = [attempts EXCEPT ![j] = @ + 1]
    /\ UNCHANGED <<pg, cache, failures, faults, req, staleServe>>

\* invalidateScript: a new generation (INCR, or a random value when it is
\* missing), then DEL the set and its marker.
InvalidateScript(j) ==
    LET k == Keys[step[j]] IN
    /\ job[j] = "running"
    /\ step[j] <= Len(Keys)
    /\ gen' = [gen EXCEPT ![k] = NewGen]
    /\ data' = [data EXCEPT ![k] = 0]
    /\ marker' = [marker EXCEPT ![k] = NoMarker]
    /\ covered' = [covered EXCEPT ![k] = pg]
    /\ step' = [step EXCEPT ![j] = @ + 1]
    /\ UNCHANGED <<pg, job, attempts, failures, faults, req, staleServe>>

Complete(j) ==
    /\ job[j] = "running"
    /\ step[j] = Len(Keys) + 1
    /\ job' = [job EXCEPT ![j] = "done"]
    /\ UNCHANGED <<pg, cache, step, attempts, failures, faults, req, staleServe>>

\* An attempt fails before its next script, or after its last one when the
\* acknowledgement is lost. It retries, or fails once attempts run out.
JobFails(j) ==
    /\ job[j] = "running"
    /\ failures < MaxFailures
    /\ failures' = failures + 1
    /\ job' = [job EXCEPT ![j] = IF attempts[j] >= MaxAttempts THEN "failed" ELSE "pending"]
    /\ UNCHANGED <<pg, cache, step, attempts, faults, req, staleServe>>

\* An operator replays the failed job as a new row.
OperatorReplay(j) ==
    /\ job[j] = "failed"
    /\ job' = [job EXCEPT ![j] = "pending"]
    /\ attempts' = [attempts EXCEPT ![j] = 0]
    /\ UNCHANGED <<pg, cache, step, failures, faults, req, staleServe>>

Invalidation(j) == Claim(j) \/ InvalidateScript(j) \/ Complete(j)

----
\* Valkey loses keys: the marker's or generation's expiry, allkeys-lru evicting
\* one key, or a restart without persistence dropping every key.

\* The marker's 28-day expiry runs out.
MarkerExpires(k) ==
    /\ Expiry
    /\ marker[k] # NoMarker
    /\ marker' = [marker EXCEPT ![k] = NoMarker]
    /\ UNCHANGED <<pg, data, gen, covered, jobVars, faults, req, staleServe>>

\* The generation's expiry runs out: an hour after a script created it, or with
\* its marker once a rebuild published under it.
GenerationExpires(k) ==
    /\ Expiry
    /\ gen[k] # NoGeneration
    /\ marker[k] = NoMarker
    /\ gen' = [gen EXCEPT ![k] = NoGeneration]
    /\ UNCHANGED <<pg, data, marker, covered, jobVars, faults, req, staleServe>>

Evict(k, part) ==
    /\ part \in Faults
    /\ faults < MaxFaults
    /\ faults' = faults + 1
    /\ data' = IF part = "data" THEN [data EXCEPT ![k] = 0] ELSE data
    /\ marker' = IF part = "marker" THEN [marker EXCEPT ![k] = NoMarker] ELSE marker
    /\ gen' = IF part = "generation" THEN [gen EXCEPT ![k] = NoGeneration] ELSE gen
    /\ UNCHANGED <<pg, covered, jobVars, req, staleServe>>

Restart ==
    /\ "restart" \in Faults
    /\ faults < MaxFaults
    /\ faults' = faults + 1
    /\ data' = [k \in KeySet |-> 0]
    /\ marker' = [k \in KeySet |-> NoMarker]
    /\ gen' = [k \in KeySet |-> NoGeneration]
    /\ UNCHANGED <<pg, covered, jobVars, req, staleServe>>

Next ==
    \/ Write
    \/ \E r \in Readers : Request(r)
    \/ \E j \in Jobs : Invalidation(j) \/ JobFails(j) \/ OperatorReplay(j)
    \/ \E k \in KeySet : MarkerExpires(k) \/ GenerationExpires(k)
    \/ \E k \in KeySet, part \in Parts : Evict(k, part)
    \/ Restart

\* Jobs keep running and, with Expiry, markers expire. With
\* OperatorReplays, an operator eventually replays a failed job. Requests,
\* writes, failures and faults are never forced.
Fairness ==
    /\ \A j \in Jobs : WF_vars(Invalidation(j))
    /\ \A k \in KeySet : WF_vars(MarkerExpires(k))
    /\ \A j \in Jobs : IF OperatorReplays THEN WF_vars(OperatorReplay(j)) ELSE TRUE

Spec == Init /\ [][Next]_vars /\ Fairness

----
\* Safety

\* A request never serves a cached board older than an invalidation that had
\* applied to that key before it read the page.
ServedFenced == ~staleServe

\* A key the cache would serve never holds a board older than the last
\* invalidation applied to it, nor a truncated set.
PublishedFenced == \A k \in KeySet : Servable(k) => data[k] >= covered[k]

\* While Valkey keeps a board's keys, a marker names the key's current
\* generation: an invalidation deletes it, and a rebuild writes only the
\* generation it checked.
MarkerNamesGeneration == \A k \in KeySet : marker[k] # NoMarker => marker[k][1] = gen[k]

\* Liveness: once writes stop, every key the cache would serve comes to hold
\* the PostgreSQL board, through completed invalidations, or through the
\* marker's expiry when no operator replays a failed job.
Fresh == \A k \in KeySet : Servable(k) => data[k] = pg

Converges == <>[]Fresh
====

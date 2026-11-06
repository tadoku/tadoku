---- MODULE LeaderboardCache ----
\* The Valkey cache of an unfiltered leaderboard: requests that serve the
\* cached board only when it was built from the board's current
\* leaderboard_versions version, and rebuild it from PostgreSQL otherwise.
\*
\* Sources:
\*   services/tadoku-api/features/leaderboard/leaderboard_cache.go
\*   services/tadoku-api/features/leaderboard/leaderboard_service.go
\*   services/tadoku-api/features/leaderboard/leaderboard_repository.go
\*   services/tadoku-api/sql/leaderboard/leaderboard.sql
\*   services/tadoku-api/migrations/0076_leaderboard_versions.up.sql
\*   services/tadoku-api/app/worker/application.go (InvalidateContestLeaderboard, InvalidateOfficialLeaderboard)
\*   services/tadoku-api/features/logs/logs_mutation_service.go
\*   services/tadoku-api/features/logs/logs_removal_service.go
\*   services/tadoku-api/features/contests/contests_service.go
\*   services/tadoku-api/features/contests/contests_removal_service.go
\*   services/tadoku-api/app/worker/account_deletion.go
\*   k8s/dev/base/data/cache.yaml (allkeys-lru, no persistence)
\*
\* The version triggers raise the board's version in the transaction of every
\* write that changes it, so a write is one step that changes the board and
\* raises its version: version v is the board after v writes. Version 0 is the
\* empty board, which Valkey stores as an absent sorted set, so a lost set
\* reads as version 0. Every other board, and a truncated set, has the same
\* size: the worst case for the size check.
\*
\* A request reads the board's version (BoardVersion, one statement), then runs
\* pageScript, which serves the page only when the ":last_updated" marker names
\* that version and the set's size. On a miss it loads the board (one statement
\* that sees the writes committed when it starts; ContestExists is folded into
\* it) and runs rebuildScript, which publishes the board under the version the
\* request read unless the marker names a later version. Reading the version
\* before the board is what keeps a published board at least as new as the
\* version its marker names. Each Lua script is one atomic step that keeps the
\* writes it made before an error. rebuildScript deletes the marker and the
\* set before it adds the scores, and only a NaN score fails after the DELs,
\* leaving a truncated set without a marker. Any Valkey error sends a request
\* to PostgreSQL, which changes no cache state. Filtered reads only read
\* PostgreSQL and are not modelled.
\*
\* Each write also enqueues an invalidation job for the previous release's
\* reads. Its invalidateScript deletes the set and the marker, like losing
\* both keys; one run per write stands for every attempt and replay.
\*
\* Faults lists what may go wrong in Valkey, and MaxFaults bounds how often:
\* eviction of one key ("data" or "marker"), since allkeys-lru evicts keys one
\* at a time, a "restart" without persistence that drops every key, a
\* "truncate", or an "error" that ends a request's Valkey step.
EXTENDS Integers, FiniteSets

CONSTANTS Readers, MaxWrites, Faults, MaxFaults

Versions == 0..MaxWrites
Truncated == -1
NoMarker == <<>>

\* A set's size as far as the marker can tell boards apart.
Size(v) == IF v = 0 THEN 0 ELSE 1

VARIABLES
    pg,             \* the board's version in PostgreSQL: committed writes
    data,           \* the version the sorted set holds; 0 when absent
    marker,         \* the ":last_updated" marker as <<version, size>>, or NoMarker
    invalidations,  \* invalidation jobs not yet run
    faults,         \* Valkey faults so far
    req,            \* each request: its next step, the version it read and the board it loaded
    staleServe      \* a request served a board older than the version it read

vars == <<pg, data, marker, invalidations, faults, req, staleServe>>
ReqIdle == [pc |-> "idle", version |-> 0, load |-> 0]

TypeOK ==
    /\ pg \in Versions
    /\ data \in Versions \cup {Truncated}
    /\ marker \in {NoMarker} \cup (Versions \X {0, 1})
    /\ invalidations \in 0..MaxWrites
    /\ faults \in 0..MaxFaults
    /\ Faults \subseteq {"data", "marker", "restart", "truncate", "error"}
    /\ \A r \in Readers : req[r].pc \in {"idle", "page", "load", "rebuild"}
    /\ staleServe \in BOOLEAN

Init ==
    /\ pg = 0
    /\ data = 0
    /\ marker = NoMarker
    /\ invalidations = 0
    /\ faults = 0
    /\ req = [r \in Readers |-> ReqIdle]
    /\ staleServe = FALSE

Set(r, rec) == req' = [req EXCEPT ![r] = rec]

\* rebuildScript may publish under version: no marker names a later one.
MayPublish(version) == IF marker = NoMarker THEN TRUE ELSE marker[1] <= version

----
\* A write to a leaderboard input commits with its version raise and its
\* invalidation job.
Write ==
    /\ pg < MaxWrites
    /\ pg' = pg + 1
    /\ invalidations' = invalidations + 1
    /\ UNCHANGED <<data, marker, faults, req, staleServe>>

----
\* Service.fetch for an unfiltered board.

\* BoardVersion.
BoardVersion(r) ==
    /\ req[r].pc = "idle"
    /\ Set(r, [ReqIdle EXCEPT !.pc = "page", !.version = pg])
    /\ UNCHANGED <<pg, data, marker, invalidations, faults, staleServe>>

\* fetchPage: pageScript serves the page, or reports a miss.
PageScript(r) ==
    /\ req[r].pc = "page"
    /\ IF marker = <<req[r].version, Size(data)>>
       THEN /\ Set(r, ReqIdle)
            /\ staleServe' = (staleServe \/ data < req[r].version)
       ELSE /\ Set(r, [req[r] EXCEPT !.pc = "load"])
            /\ UNCHANGED staleServe
    /\ UNCHANGED <<pg, data, marker, invalidations, faults>>

\* ContestLeaderboard or OfficialLeaderboard.
Load(r) ==
    /\ req[r].pc = "load"
    /\ Set(r, [req[r] EXCEPT !.pc = "rebuild", !.load = pg])
    /\ UNCHANGED <<pg, data, marker, invalidations, faults, staleServe>>

\* rebuildScript. The request serves its PostgreSQL board either way.
RebuildScript(r) ==
    /\ req[r].pc = "rebuild"
    /\ IF MayPublish(req[r].version)
       THEN /\ data' = req[r].load
            /\ marker' = <<req[r].version, Size(req[r].load)>>
       ELSE UNCHANGED <<data, marker>>
    /\ Set(r, ReqIdle)
    /\ UNCHANGED <<pg, invalidations, faults, staleServe>>

\* A NaN score stops rebuildScript after its DELs and some chunks.
RebuildScriptTruncates(r) ==
    /\ req[r].pc = "rebuild"
    /\ "truncate" \in Faults
    /\ faults < MaxFaults
    /\ MayPublish(req[r].version)
    /\ faults' = faults + 1
    /\ data' = Truncated
    /\ marker' = NoMarker
    /\ Set(r, ReqIdle)
    /\ UNCHANGED <<pg, invalidations, staleServe>>

\* A Valkey error or timeout: the request reads PostgreSQL instead. A rebuild
\* that errors after its script applied is RebuildScript.
ValkeyError(r) ==
    /\ req[r].pc \in {"page", "rebuild"}
    /\ "error" \in Faults
    /\ faults < MaxFaults
    /\ faults' = faults + 1
    /\ Set(r, ReqIdle)
    /\ UNCHANGED <<pg, data, marker, invalidations, staleServe>>

Request(r) == BoardVersion(r) \/ PageScript(r) \/ Load(r) \/ RebuildScript(r)

----
\* The worker's InvalidateContestLeaderboard or InvalidateOfficialLeaderboard.
InvalidateScript ==
    /\ invalidations > 0
    /\ invalidations' = invalidations - 1
    /\ data' = 0
    /\ marker' = NoMarker
    /\ UNCHANGED <<pg, faults, req, staleServe>>

----
\* Valkey loses keys: allkeys-lru evicting one key, or a restart without
\* persistence dropping every key.

Evict(part) ==
    /\ part \in Faults
    /\ faults < MaxFaults
    /\ faults' = faults + 1
    /\ data' = IF part = "data" THEN 0 ELSE data
    /\ marker' = IF part = "marker" THEN NoMarker ELSE marker
    /\ UNCHANGED <<pg, invalidations, req, staleServe>>

Restart ==
    /\ "restart" \in Faults
    /\ faults < MaxFaults
    /\ faults' = faults + 1
    /\ data' = 0
    /\ marker' = NoMarker
    /\ UNCHANGED <<pg, invalidations, req, staleServe>>

Next ==
    \/ Write
    \/ \E r \in Readers : Request(r) \/ RebuildScriptTruncates(r) \/ ValkeyError(r)
    \/ InvalidateScript
    \/ \E part \in {"data", "marker"} : Evict(part)
    \/ Restart

\* Readers keep sending requests. Writes, invalidations and faults are never
\* forced.
Fairness == \A r \in Readers : WF_vars(Request(r))

Spec == Init /\ [][Next]_vars /\ Fairness

----
\* Safety

\* A request never serves a board older than the version it read before the
\* page, so it sees every write committed before that read.
FreshOnServe == ~staleServe

\* A marker that matches its set names a version no newer than the board the
\* set holds, and never a truncated set.
PublishedFresh == (marker # NoMarker /\ marker[2] = Size(data)) => data >= marker[1]

\* While Valkey keeps the marker, a rebuild never replaces it with an older
\* version.
NoRegress == [][(marker # NoMarker /\ marker' # NoMarker) => marker'[1] >= marker[1]]_vars

\* Liveness: once writes stop, the cache comes to hold the PostgreSQL board
\* under its version and keeps it.
Converges == <>[](marker = <<pg, Size(pg)>> /\ data = pg)
====

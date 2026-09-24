---
sidebar_position: 1
title: Account deletion
description: How to manually delete a Tadoku account from the Tadoku API database, the leaderboard cache, Keto and Kratos.
---

# Account deletion

Read this when a user has asked for their Tadoku account to be deleted.

Account deletion is manual. Run the steps below in order. You can safely rerun any step.

The deletion removes the account's logs, contest registrations, roles and profile
rows. The `users` row stays as an anonymised tombstone: contest queries join
contest owners to `users`, and the Tadoku API rejects writes and profile
synchronisation for users whose `deletion_locked_at` or `deleted_at` is set.

## Prerequisites

- Write access to the Tadoku API database through `psql`. The statements use
  unqualified table names, like the migrations in
  `services/tadoku-api/migrations/`, so connect with the `search_path` that the
  Tadoku API uses.
- Access to the Kratos admin API and the Keto read and write APIs.
- The account ID. This is the Kratos identity ID, which is also `users.id`. To
  find it from an email address, run
  `curl "$KRATOS_ADMIN_URL/admin/identities?credentials_identifier=<EMAIL>"`.

Set the ID once for `psql` and once for the shell:

```sql
\set account_id '<ACCOUNT_ID>'
```

```bash
ACCOUNT_ID='<ACCOUNT_ID>'
```

## 1. Preflight

Run these read-only queries. First, list the data the account owns:

```sql
select
  (select count(*) from users where id = :'account_id') as users,
  (select deletion_locked_at from users where id = :'account_id') as deletion_locked_at,
  (select deleted_at from users where id = :'account_id') as deleted_at,
  (select count(*) from logs where user_id = :'account_id') as logs,
  (select count(*) from contest_registrations where user_id = :'account_id') as registrations,
  (select count(*) from contests where owner_user_id = :'account_id') as owned_contests,
  (select count(*) from user_roles where user_id = :'account_id') as user_roles,
  (select count(*) from profiles where user_id = :'account_id') as profiles,
  (select count(*) from account_deletion_requests where identity_id = :'account_id') as deletion_requests,
  (select count(*) from moderation_audit_log
    where user_id = :'account_id' or metadata->>'target_user_id' = :'account_id') as audit_entries;
```

If `users` is `0`, the account has no Tadoku API data. Skip to
[Keto relations](#5-keto-relations-and-feature-access).

Next, check for logs in finished contests:

```sql
select contests.id, contests.title, contests.contest_end, count(*) as logs
from contest_logs
join logs on logs.id = contest_logs.log_id
join contests on contests.id = contest_logs.contest_id
where logs.user_id = :'account_id'
  and logs.deleted_at is null
  and contests.deleted_at is null
  and contests.contest_end < current_date
group by contests.id;
```

**If this query returns any rows, stop.** Deleting these logs would change the
results of contests that have already finished. This runbook does not cover
keeping them. Do not change the account.

Finally, review the contests that the account owns:

```sql
select id, title, contest_start, contest_end, "private", official, deleted_at
from contests
where owner_user_id = :'account_id';
```

The delete step keeps owned contests and shows them with the organiser
"Deleted organizer". Contest titles and descriptions are not changed.

## 2. Lock the account

Commit the lock on its own before you delete anything. Once it is committed, the
Tadoku API rejects every log, contest and registration write for this account
with `409` and `{"error":"account_deletion_in_progress"}`.

```sql
update users
set deletion_locked_at = coalesce(deletion_locked_at, now())
where id = :'account_id';
```

## 3. Delete the data

Run this transaction in the same `psql` session. The outbox inserts must run
before the deletes because they read the rows that the deletes remove.

```sql
begin;

insert into leaderboard_outbox (event_type, user_id, contest_id)
select 'refresh_contest_score', :'account_id'::uuid, contest_id
from (
  select contest_id from contest_registrations where user_id = :'account_id'
  union
  select contest_logs.contest_id
  from contest_logs
  join logs on logs.id = contest_logs.log_id
  where logs.user_id = :'account_id'
) as affected_contests;

insert into leaderboard_outbox (event_type, user_id, year)
select distinct 'refresh_official_scores', :'account_id'::uuid, year
from logs
where user_id = :'account_id'
  and eligible_official_leaderboard;

delete from contest_logs
using logs
where logs.id = contest_logs.log_id
  and logs.user_id = :'account_id';

-- log_tags rows are removed by their on delete cascade foreign key.
delete from logs where user_id = :'account_id';

delete from contest_registrations where user_id = :'account_id';
delete from user_roles where user_id = :'account_id';
delete from profiles where user_id = :'account_id';
delete from account_deletion_requests where identity_id = :'account_id';

update contests
set owner_user_display_name = 'Deleted organizer'
where owner_user_id = :'account_id';

update users
set
  display_name = 'Deleted participant',
  deleted_at = coalesce(deleted_at, now()),
  updated_at = now()
where id = :'account_id'
  and deletion_locked_at is not null;
```

Check the last statement's output. If it printed `UPDATE 1`, commit:

```sql
commit;
```

If it printed `UPDATE 0`, the lock is missing. Run `rollback;` and go back to
step 2.

Moderation audit entries stay in `moderation_audit_log`, both entries where the
account was the moderator and entries where it was the target.

## 4. Leaderboards

The Tadoku API leaderboard outbox worker processes the rows inserted in step 3.
It invalidates the cached contest, yearly and global leaderboards in Valkey. On
the next read, each leaderboard is rebuilt from the Tadoku API database. The
rebuilt leaderboards leave out the account, because its registrations and logs
are gone and official leaderboards skip users with `deleted_at` set. You do not
need to change the cache by hand.

Wait a few seconds, then run this query. It returns `0` once the worker has
processed the rows:

```sql
select count(*)
from leaderboard_outbox
where user_id = :'account_id'
  and processed_at is null;
```

## 5. Keto relations and feature access

Tadoku stores roles as `app:tadoku#admins` and `app:tadoku#banned` tuples. Remove
both through the Keto write API. The Verify section checks that none are left.

```bash
for relation in admins banned; do
  curl -sS -X DELETE \
    "$KETO_WRITE_URL/admin/relation-tuples?namespace=app&object=tadoku&relation=$relation&subject_id=$ACCOUNT_ID"
done
```

Flipt stores named-user feature access by account ID. The flags that support it
are listed in `services/tadoku-api/features/featureflags/domain.go`. For each of
those flags, call `GET /immersion/admin/feature-flags/{flagKey}/users/{userId}`
as an administrator. If the account has access, send `DELETE` to the same path
to revoke it.

## 6. Kratos identity

Delete the identity through the Kratos admin API:

```bash
curl -fsS -X DELETE "$KRATOS_ADMIN_URL/admin/identities/$ACCOUNT_ID"
```

A `404` response means the identity is already gone.

## Verify

In the Tadoku API database:

```sql
select display_name, deletion_locked_at, deleted_at from users where id = :'account_id';

select
  (select count(*) from logs where user_id = :'account_id') as logs,
  (select count(*) from log_tags where user_id = :'account_id') as log_tags,
  (select count(*) from contest_registrations where user_id = :'account_id') as registrations,
  (select count(*) from user_roles where user_id = :'account_id') as user_roles,
  (select count(*) from profiles where user_id = :'account_id') as profiles,
  (select count(*) from account_deletion_requests where identity_id = :'account_id') as deletion_requests,
  (select count(*) from contests
    where owner_user_id = :'account_id'
      and owner_user_display_name <> 'Deleted organizer') as named_contests;
```

The `users` query returns one row, with `Deleted participant` as the display name
and both timestamps set. Every count in the second query is `0`.

In Keto and Kratos:

```bash
# Expect "relation_tuples":[]
curl -fsS "$KETO_READ_URL/relation-tuples?namespace=app&subject_id=$ACCOUNT_ID"

# Expect 404
curl -s -o /dev/null -w '%{http_code}\n' "$KRATOS_ADMIN_URL/admin/identities/$ACCOUNT_ID"
```

The admin user list reads identities from an in-memory Kratos cache. The account
disappears from that list within five minutes.

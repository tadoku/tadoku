begin;

create table async_outbox (
  id bigserial primary key,
  task_type text not null,
  payload jsonb not null,
  state text not null default 'pending',
  attempts integer not null default 0,
  created_at timestamptz not null default now(),
  next_attempt_at timestamptz not null default now(),
  claim_token uuid,
  lease_expires_at timestamptz,
  completed_at timestamptz,
  failed_at timestamptz,
  last_error text,
  replay_of_id bigint references async_outbox (id) on delete restrict,
  replay_actor text,
  replay_reason text,

  constraint async_outbox_task_type_versioned
    check (task_type ~ '^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)*\.v[1-9][0-9]*$'),
  constraint async_outbox_payload_object
    check (jsonb_typeof(payload) = 'object'),
  constraint async_outbox_state_valid
    check (state in ('pending', 'running', 'completed', 'failed')),
  constraint async_outbox_attempts_nonnegative
    check (attempts >= 0),
  constraint async_outbox_claim_running
    check (
      (state = 'running' and claim_token is not null and lease_expires_at is not null)
      or (state <> 'running' and claim_token is null and lease_expires_at is null)
    ),
  constraint async_outbox_terminal_timestamps
    check (
      (state = 'completed' and completed_at is not null and failed_at is null)
      or (state = 'failed' and completed_at is null and failed_at is not null)
      or (state in ('pending', 'running') and completed_at is null and failed_at is null)
    ),
  constraint async_outbox_last_error_redacted
    check (last_error is null or (length(last_error) <= 100 and last_error ~ '^[a-z0-9_]+$')),
  constraint async_outbox_replay_linked
    check (
      (replay_of_id is null and replay_actor is null and replay_reason is null)
      or (
        replay_of_id is not null
        and replay_of_id <> id
        and replay_actor is not null
        and length(btrim(replay_actor)) between 1 and 200
        and replay_reason is not null
        and length(btrim(replay_reason)) between 1 and 500
      )
    )
);

create index async_outbox_pending_due
  on async_outbox (task_type, next_attempt_at, id)
  where state = 'pending';

create index async_outbox_expired_claims
  on async_outbox (lease_expires_at, id)
  where state = 'running';

create index async_outbox_completed_retention
  on async_outbox (completed_at, id)
  where state = 'completed';

create index async_outbox_replay_of_id
  on async_outbox (replay_of_id)
  where replay_of_id is not null;

commit;

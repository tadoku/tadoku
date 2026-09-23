begin;

create table profiles (
  user_id uuid primary key not null,

  created_at timestamp not null default now(),
  updated_at timestamp not null default now()
);

create table account_deletion_requests (
  id uuid primary key default uuid_generate_v4(),
  identity_id uuid not null unique,
  status varchar(32) not null,
  resume_status varchar(32),

  accepted_at timestamp not null,
  discord_channel_id varchar(32),
  discord_message_id varchar(32),

  queued_at timestamp,
  access_locked_at timestamp,
  immersion_scrubbed_at timestamp,
  caches_reconciled_at timestamp,
  authorization_removed_at timestamp,
  identity_deleted_at timestamp,
  completed_at timestamp,

  attempt_count integer not null default 0,
  next_attempt_at timestamp,
  last_error_code varchar(100),
  lease_owner uuid,
  lease_expires_at timestamp,
  lease_generation bigint not null default 0,

  manual_attention_at timestamp,
  remediation_due_at timestamp,

  created_at timestamp not null default now(),
  updated_at timestamp not null default now(),

  constraint account_deletion_requests_status_valid
    check (status in (
      'receipt_pending',
      'queued',
      'access_locked',
      'immersion_scrubbed',
      'caches_reconciled',
      'authorization_removed',
      'identity_deleted',
      'complete',
      'manual_attention'
    )),
  constraint account_deletion_requests_resume_status_valid
    check (
      (
        status = 'manual_attention'
        and resume_status in (
          'queued',
          'access_locked',
          'immersion_scrubbed',
          'caches_reconciled',
          'authorization_removed',
          'identity_deleted'
        )
      )
      or (status <> 'manual_attention' and resume_status is null)
    ),
  constraint account_deletion_requests_receipt_valid
    check (
      status = 'receipt_pending'
      or (discord_channel_id is not null and discord_message_id is not null)
    ),
  constraint account_deletion_requests_receipt_ids_paired
    check (
      (discord_channel_id is null and discord_message_id is null)
      or (discord_channel_id is not null and discord_message_id is not null)
    ),
  constraint account_deletion_requests_attempt_count_nonnegative
    check (attempt_count >= 0),
  constraint account_deletion_requests_error_code_sanitized
    check (last_error_code is null or last_error_code ~ '^[a-z0-9_]+$'),
  constraint account_deletion_requests_lease_paired
    check (
      (lease_owner is null and lease_expires_at is null)
      or (lease_owner is not null and lease_expires_at is not null)
    ),
  constraint account_deletion_requests_lease_generation_nonnegative
    check (lease_generation >= 0),
  constraint account_deletion_requests_manual_attention_valid
    check (
      (
        status = 'manual_attention'
        and manual_attention_at is not null
        and remediation_due_at = manual_attention_at + interval '7 days'
      )
      or (
        status <> 'manual_attention'
        and manual_attention_at is null
        and remediation_due_at is null
      )
    )
);

create index account_deletion_requests_worker_claim
  on account_deletion_requests (next_attempt_at, lease_expires_at, created_at)
  where status in (
    'queued',
    'access_locked',
    'immersion_scrubbed',
    'caches_reconciled',
    'authorization_removed',
    'identity_deleted'
  );

create index account_deletion_requests_manual_attention_due
  on account_deletion_requests (remediation_due_at)
  where status = 'manual_attention';

commit;

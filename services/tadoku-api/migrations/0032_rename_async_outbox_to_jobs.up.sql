begin;

alter table async_outbox rename to jobs;
alter sequence async_outbox_id_seq rename to jobs_id_seq;

alter table jobs rename constraint async_outbox_pkey to jobs_pkey;
alter table jobs rename constraint async_outbox_replay_of_id_fkey to jobs_replay_of_id_fkey;
alter table jobs rename constraint async_outbox_task_type_versioned to jobs_task_type_versioned;
alter table jobs rename constraint async_outbox_payload_object to jobs_payload_object;
alter table jobs rename constraint async_outbox_state_valid to jobs_state_valid;
alter table jobs rename constraint async_outbox_attempts_nonnegative to jobs_attempts_nonnegative;
alter table jobs rename constraint async_outbox_claim_running to jobs_claim_running;
alter table jobs rename constraint async_outbox_terminal_timestamps to jobs_terminal_timestamps;
alter table jobs rename constraint async_outbox_last_error_redacted to jobs_last_error_redacted;
alter table jobs rename constraint async_outbox_replay_linked to jobs_replay_linked;

alter index async_outbox_pending_due rename to jobs_pending_due;
alter index async_outbox_expired_claims rename to jobs_expired_claims;
alter index async_outbox_completed_retention rename to jobs_completed_retention;
alter index async_outbox_replay_of_id rename to jobs_replay_of_id;

commit;

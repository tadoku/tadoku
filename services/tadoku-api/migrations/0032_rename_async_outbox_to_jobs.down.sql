begin;

alter table jobs rename to async_outbox;
alter sequence jobs_id_seq rename to async_outbox_id_seq;

alter table async_outbox rename constraint jobs_pkey to async_outbox_pkey;
alter table async_outbox rename constraint jobs_replay_of_id_fkey to async_outbox_replay_of_id_fkey;
alter table async_outbox rename constraint jobs_task_type_versioned to async_outbox_task_type_versioned;
alter table async_outbox rename constraint jobs_payload_object to async_outbox_payload_object;
alter table async_outbox rename constraint jobs_state_valid to async_outbox_state_valid;
alter table async_outbox rename constraint jobs_attempts_nonnegative to async_outbox_attempts_nonnegative;
alter table async_outbox rename constraint jobs_claim_running to async_outbox_claim_running;
alter table async_outbox rename constraint jobs_terminal_timestamps to async_outbox_terminal_timestamps;
alter table async_outbox rename constraint jobs_last_error_redacted to async_outbox_last_error_redacted;
alter table async_outbox rename constraint jobs_replay_linked to async_outbox_replay_linked;

alter index jobs_pending_due rename to async_outbox_pending_due;
alter index jobs_expired_claims rename to async_outbox_expired_claims;
alter index jobs_completed_retention rename to async_outbox_completed_retention;
alter index jobs_replay_of_id rename to async_outbox_replay_of_id;

commit;

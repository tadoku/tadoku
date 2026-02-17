create table leaderboard_outbox (
  id bigserial primary key,
  event_type text not null,
  user_id uuid not null,
  contest_id uuid,
  year smallint,
  created_at timestamp not null default now(),
  processed_at timestamp
);

create index leaderboard_outbox_unprocessed on leaderboard_outbox (id) where processed_at is null;

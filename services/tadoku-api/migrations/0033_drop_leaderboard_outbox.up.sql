begin;

lock table leaderboard_outbox in access exclusive mode;

do $$
begin
  if exists (select 1 from leaderboard_outbox where processed_at is null) then
    raise exception 'cannot drop leaderboard_outbox while unprocessed rows remain';
  end if;
end $$;

drop table leaderboard_outbox;

commit;

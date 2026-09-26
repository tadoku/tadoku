select (select count(*)::text from leaderboard_outbox) as legacy_total,
  (select count(*)::text from leaderboard_outbox where processed_at is null) as legacy_pending,
  (select count(*)::text from jobs) as async_total
order by 1;

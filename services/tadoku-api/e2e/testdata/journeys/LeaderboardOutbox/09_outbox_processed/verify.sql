select count(*)::text as pending from leaderboard_outbox where processed_at is null;

select
  (select json_agg(contest_registrations.contest_id order by contest_registrations.contest_id)
    from contest_registrations
    where contest_registrations.tenant = 'tadoku/prod'
      and contest_registrations.user_id = '11111111-1111-4111-8111-111111111111') as registrations,
  (select json_agg(json_build_object(
      'id', logs.id,
      'description', logs.description,
      'frozen', logs.frozen_at is not null,
      'contests', (select json_agg(contest_logs.contest_id order by contest_logs.contest_id)
        from contest_logs
        where contest_logs.log_id = logs.id)
    ) order by logs.id)
    from logs
    where logs.tenant = 'tadoku/prod'
      and logs.user_id = '11111111-1111-4111-8111-111111111111') as logs,
  (select json_agg(jobs.payload order by jobs.id)
    from jobs
    where jobs.tenant = 'tadoku/prod'
      and jobs.task_type like 'leaderboard.%') as leaderboard_jobs

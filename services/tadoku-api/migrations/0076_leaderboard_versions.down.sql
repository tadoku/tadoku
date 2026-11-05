begin;

set local lock_timeout = '5s';

drop trigger leaderboard_versions_contests on contests;
drop trigger leaderboard_versions_users on users;
drop trigger leaderboard_versions_contest_registrations on contest_registrations;
drop trigger leaderboard_versions_contest_logs on contest_logs;
drop trigger leaderboard_versions_logs on logs;

drop function leaderboard_versions_from_contests();
drop function leaderboard_versions_from_users();
drop function leaderboard_versions_from_contest_registrations();
drop function leaderboard_versions_from_contest_logs();
drop function leaderboard_versions_from_logs();
drop function raise_leaderboard_versions(text, text[]);

drop table leaderboard_versions;

commit;

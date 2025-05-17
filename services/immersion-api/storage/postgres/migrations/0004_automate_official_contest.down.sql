-- Drop helper and main function
drop function if exists data.create_contest_round(int, int);
drop function if exists data.create_official_contest(date, date, int);
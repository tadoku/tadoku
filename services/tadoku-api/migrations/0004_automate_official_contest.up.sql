-- Create main contest function
create function create_official_contest(contest_start date, contest_end date, contest_number int)
  returns uuid
  language plpgsql
as
$$
declare
  owner_id uuid;
  owner_display_name varchar(255);
  registration_end date := (contest_end - interval '1 week')::date;
  title varchar(255) := format('%s Round %s', extract(year from contest_start)::int, contest_number);
  result_id uuid;
begin
  -- Constraints
  if contest_start >= contest_end then
    raise exception 'contest_start (%) must be before contest_end (%)', contest_start, contest_end;
  end if;

  if contest_number < 1 then
    raise exception 'contest_number (%) must be a positive integer', contest_number;
  end if;

  -- Lookup admin user
  select id, display_name
  into owner_id, owner_display_name
  from users
  where display_name = 'antonve'
  limit 1;

  if not found then
    raise exception 'Admin user "antonve" not found';
  end if;

  raise notice 'Creating contest: %', title;

  -- Insert contest
  begin
    insert into contests (
      owner_user_id, owner_user_display_name, private,
      contest_start, contest_end, registration_end,
      title, activity_type_id_allow_list, official
    )
    values (
      owner_id, owner_display_name, false,
      contest_start, contest_end, registration_end,
      title, array[1, 2], true
    )
    returning id into result_id;

    raise notice 'Contest inserted successfully with id %', result_id;
  exception
    when others then
      raise warning 'Failed to insert contest: %', SQLERRM;
      return null;
  end;

  return result_id;
end;
$$;

-- Create helper to generate contest rounds
create function create_contest_round(round_number int, contest_year int)
returns void as
$$
declare
  contest_start date;
  contest_end date;
begin
  case round_number
    when 1 then
      contest_start := make_date(contest_year, 1, 1);
      contest_end := make_date(contest_year, 1, 31);
    when 2 then
      contest_start := make_date(contest_year, 3, 1);
      contest_end := make_date(contest_year, 3, 31);
    when 3 then
      contest_start := make_date(contest_year, 5, 1);
      contest_end := make_date(contest_year, 5, 14);
    when 4 then
      contest_start := make_date(contest_year, 7, 1);
      contest_end := make_date(contest_year, 7, 31);
    when 5 then
      contest_start := make_date(contest_year, 9, 1);
      contest_end := make_date(contest_year, 9, 30);
    when 6 then
      contest_start := make_date(contest_year, 11, 1);
      contest_end := make_date(contest_year, 11, 14);
    else
      raise exception 'Invalid round number: %', round_number;
  end case;

  perform create_official_contest(contest_start, contest_end, round_number);
end;
$$ language plpgsql;
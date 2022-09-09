-- creates a generated column "year" based on the creation date of the record
alter table contest_logs add column "year" smallint generated always as (extract(year from created_at)) stored;

-- index the year so we can filter effectively on the year
create index contest_logs_year on contest_logs(year);

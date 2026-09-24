-- Reuse the shared development users, but write fixtures only into this branch.
\getenv branch_database PGDATABASE
select current_database() = :'branch_database'
       and current_user = 'tadoku'
       and pg_get_userbyid(datdba) = 'tadoku'
       and octet_length(:'branch_database') <= 63
       and :'branch_database' ~ '^tadoku-[a-z0-9][a-z0-9-]*-[0-9a-f]{8}$'
       and shobj_description(oid, 'pg_database') =
           'dev-cli branch database route=' || substring(:'branch_database' from 8)
       as owned
from pg_database where datname = current_database()
\gset
\if :owned
\else
  do $$ begin raise exception 'refusing seed outside an owned branch database'; end $$;
\endif

\connect tadoku
select count(*) = 1 as admin_found, min(id::text) as admin_user_id
from users where display_name = 'Dev Admin'
\gset
select count(*) = 1 as reader_found, min(id::text) as reader_user_id
from users where display_name = 'Dev Reader'
\gset
\if :admin_found
\else
  do $$ begin raise exception 'expected one shared Dev Admin fixture; run make dev-seed first'; end $$;
\endif
\if :reader_found
\else
  do $$ begin raise exception 'expected one shared Dev Reader fixture; run make dev-seed first'; end $$;
\endif

\connect :branch_database
\ir /seed/immersion.sql
\ir /seed/profile.sql
\ir /seed/content.sql

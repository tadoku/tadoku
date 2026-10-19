#!/bin/sh
set -eu

if [ "${PGHOST:-}" != tadoku-dev-db.tdk-dev-data ]; then
  echo 'refusing tenant provisioning outside the development Postgres service' >&2
  exit 1
fi
if [ "${DEV_TENANT:-}" != "tadoku/${DEV_ROUTE:?route is required}" ]; then
  echo 'refusing a tenant that differs from the development route' >&2
  exit 1
fi

fixture_ids=$(psql -X -q -At --dbname=tadoku --set=ON_ERROR_STOP=1 <<'SQL'
begin;
select set_config('tadoku.tenant', 'tadoku/prod', true) as tenant \gset
select count(*) = 1 as admin_found, min(id::text) as admin_user_id
from users
where tenant = 'tadoku/prod' and display_name = 'Dev Admin'
\gset
select count(*) = 1 as reader_found, min(id::text) as reader_user_id
from users
where tenant = 'tadoku/prod' and display_name = 'Dev Reader'
\gset
\if :admin_found
\else
  do $$ begin raise exception 'expected one base Dev Admin fixture; authorized shared setup is required'; end $$;
\endif
\if :reader_found
\else
  do $$ begin raise exception 'expected one base Dev Reader fixture; authorized shared setup is required'; end $$;
\endif
\echo :admin_user_id :reader_user_id
commit;
SQL
)
set -- $fixture_ids
if [ "$#" -ne 2 ]; then
  echo 'expected base development fixture IDs; authorized shared setup is required' >&2
  exit 1
fi

/tadoku-tenant provision --tenant "$DEV_TENANT" --flipt-features /flipt/features.yaml --tester "$2"
TENANT_PRODUCTION=false /tadoku-tenant development-accounts --tenant "$DEV_TENANT" \
  --admin-id "$1" --reader-id "$2"
exec psql -X --set=ON_ERROR_STOP=1 \
  --set=tenant="$DEV_TENANT" --set=admin_user_id="$1" --set=reader_user_id="$2" \
  --file=/seed.sql

\if :{?tenant}
\else
  do $$ begin raise exception 'tenant psql variable is required'; end $$;
\endif

begin;
select set_config('tadoku.tenant', :'tenant', true);
select current_user = 'tadoku'
       and :'tenant' ~ '^tadoku/[a-z0-9][a-z0-9-]*-[0-9a-f]{8}$'
       and exists(select from tenants where key = :'tenant' and kind = 'test') as allowed
\gset
\if :allowed
\else
  do $$ begin raise exception 'refusing seed outside a registered branch tenant with the runtime role'; end $$;
\endif
commit;

\ir /seed/immersion.sql
\ir /seed/profile.sql
\ir /seed/content.sql

\if :{?tenant}
\else
  do $$ begin raise exception 'tenant psql variable is required'; end $$;
\endif

begin;
select set_config('tadoku.tenant', :'tenant', true);

select
  case when :'tenant' = 'tadoku/prod' then '00000000-0000-4000-8000-000000000401'::uuid
       else uuid_generate_v5('00000000-0000-4000-8000-000000000401'::uuid, :'tenant') end as page_id,
  case when :'tenant' = 'tadoku/prod' then '00000000-0000-4000-8000-000000000402'::uuid
       else uuid_generate_v5('00000000-0000-4000-8000-000000000402'::uuid, :'tenant') end as page_content_id,
  case when :'tenant' = 'tadoku/prod' then '00000000-0000-4000-8000-000000000501'::uuid
       else uuid_generate_v5('00000000-0000-4000-8000-000000000501'::uuid, :'tenant') end as post_id,
  case when :'tenant' = 'tadoku/prod' then '00000000-0000-4000-8000-000000000502'::uuid
       else uuid_generate_v5('00000000-0000-4000-8000-000000000502'::uuid, :'tenant') end as post_content_id,
  case when :'tenant' = 'tadoku/prod' then '00000000-0000-4000-8000-000000000601'::uuid
       else uuid_generate_v5('00000000-0000-4000-8000-000000000601'::uuid, :'tenant') end as announcement_id
\gset

insert into pages (
  tenant,
  id,
  "namespace",
  slug,
  current_content_id,
  published_at,
  created_at,
  updated_at
)
values
  (
    :'tenant',
    :'page_id'::uuid,
    'tadoku',
    'dev-welcome',
    :'page_content_id'::uuid,
    now(),
    now(),
    now()
  )
on conflict (tenant, "namespace", slug) do update
set
  current_content_id = excluded.current_content_id,
  published_at = excluded.published_at,
  updated_at = now(),
  deleted_at = null;

insert into pages_content (
  tenant,
  id,
  page_id,
  title,
  html,
  created_at
)
values
  (
    :'tenant',
    :'page_content_id'::uuid,
    :'page_id'::uuid,
    'Dev Welcome',
    '<p>This page is seeded by the Tadoku dev environment.</p>',
    now()
  )
on conflict (id) do update
set
  page_id = excluded.page_id,
  title = excluded.title,
  html = excluded.html;

insert into posts (
  tenant,
  id,
  "namespace",
  slug,
  current_content_id,
  published_at,
  created_at,
  updated_at
)
values
  (
    :'tenant',
    :'post_id'::uuid,
    'tadoku',
    'dev-round-open',
    :'post_content_id'::uuid,
    now(),
    now(),
    now()
  )
on conflict (tenant, "namespace", slug) do update
set
  current_content_id = excluded.current_content_id,
  published_at = excluded.published_at,
  updated_at = now(),
  deleted_at = null;

insert into posts_content (
  tenant,
  id,
  post_id,
  title,
  content,
  created_at
)
values
  (
    :'tenant',
    :'post_content_id'::uuid,
    :'post_id'::uuid,
    'Dev Round Is Open',
    'This seeded post gives the dev site a small content dataset.',
    now()
  )
on conflict (id) do update
set
  post_id = excluded.post_id,
  title = excluded.title,
  content = excluded.content;

insert into announcements (
  tenant,
  id,
  "namespace",
  title,
  content,
  style,
  href,
  starts_at,
  ends_at,
  created_at,
  updated_at
)
values
  (
    :'tenant',
    :'announcement_id'::uuid,
    'tadoku',
    'Seeded dev data',
    'The dev database has been migrated and seeded.',
    'info',
    '/pages/dev-welcome',
    now() - interval '1 day',
    now() + interval '30 days',
    now(),
    now()
  )
on conflict (id) do update
set
  title = excluded.title,
  content = excluded.content,
  style = excluded.style,
  href = excluded.href,
  starts_at = excluded.starts_at,
  ends_at = excluded.ends_at,
  updated_at = now(),
  deleted_at = null;

commit;

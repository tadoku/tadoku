begin;

insert into pages (
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
    '00000000-0000-4000-8000-000000000401',
    'tadoku',
    'dev-welcome',
    '00000000-0000-4000-8000-000000000402',
    now(),
    now(),
    now()
  )
on conflict ("namespace", slug) do update
set
  current_content_id = excluded.current_content_id,
  published_at = excluded.published_at,
  updated_at = now(),
  deleted_at = null;

insert into pages_content (
  id,
  page_id,
  title,
  html,
  created_at
)
values
  (
    '00000000-0000-4000-8000-000000000402',
    '00000000-0000-4000-8000-000000000401',
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
    '00000000-0000-4000-8000-000000000501',
    'tadoku',
    'dev-round-open',
    '00000000-0000-4000-8000-000000000502',
    now(),
    now(),
    now()
  )
on conflict ("namespace", slug) do update
set
  current_content_id = excluded.current_content_id,
  published_at = excluded.published_at,
  updated_at = now(),
  deleted_at = null;

insert into posts_content (
  id,
  post_id,
  title,
  content,
  created_at
)
values
  (
    '00000000-0000-4000-8000-000000000502',
    '00000000-0000-4000-8000-000000000501',
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
    '00000000-0000-4000-8000-000000000601',
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

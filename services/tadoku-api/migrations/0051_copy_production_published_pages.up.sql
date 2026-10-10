-- Gives a new test tenant production's currently published CMS pages.
-- tadoku-tenant provision calls it after inserting the test registry row; the
-- infrastructure grants execute to tadoku_tenant_lifecycle.
begin;

set local lock_timeout = '5s';

select set_config('search_path', format('%I, pg_temp', current_schema()), true);

create function copy_production_published_pages(target_tenant text)
  returns void
  language plpgsql
  security definer
  set search_path from current
as
$$
begin
  if not exists (
    select
    from tenants
    where key = target_tenant
      and kind = 'test'
  ) then
    raise exception 'tenant % is not a registered test tenant', target_tenant;
  end if;

  if exists (select from pages where tenant = target_tenant) then
    return;
  end if;

  with published as materialized (
    select
      gen_random_uuid() as page_id,
      gen_random_uuid() as content_id,
      pages."namespace",
      pages.slug,
      pages.published_at,
      pages.created_at,
      pages.updated_at,
      pages_content.title,
      pages_content.html,
      pages_content.created_at as content_created_at
    from pages
    inner join pages_content on pages_content.id = pages.current_content_id
    where pages.tenant = 'tadoku/prod'
      and pages_content.tenant = 'tadoku/prod'
      and pages.deleted_at is null
      and pages.published_at <= now() at time zone 'utc'
  ),
  copied_pages as (
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
    select target_tenant, page_id, "namespace", slug, content_id, published_at, created_at, updated_at
    from published
  )
  insert into pages_content (tenant, id, page_id, title, html, created_at)
  select target_tenant, content_id, page_id, title, html, content_created_at
  from published;
end;
$$;

revoke execute on function copy_production_published_pages(text) from public;

commit;

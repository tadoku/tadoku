begin;

-- Remove the retired forum option from the contact page. Other contact
-- options stay. Databases without that page, or without this wording, are unchanged.
with removed as (
  update pages_content as content
  set html = regexp_replace(
    content.html,
    '[[:space:]]*<li>[[:space:]]*Make a post on the [Ff]orum[[:space:]]*</li>',
    '',
    'g'
  )
  from pages
  where content.page_id = pages.id
    and pages."namespace" = 'tadoku'
    and pages.slug = 'contact'
    and content.html ~* 'Make a post on the Forum'
  returning content.page_id
)
update pages
set updated_at = now()
where id in (select page_id from removed);

commit;

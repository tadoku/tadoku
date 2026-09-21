begin;

with restored as (
  update pages_content as content
  set html = regexp_replace(
    content.html,
    '</ul>',
    '<li>Make a post on the Forum</li></ul>'
  )
  from pages
  where content.page_id = pages.id
    and pages."namespace" = 'tadoku'
    and pages.slug = 'contact'
    and content.html !~* 'Make a post on the Forum'
    and content.html like '%</ul>%'
  returning content.page_id
)
update pages
set updated_at = now()
where id in (select page_id from restored);

commit;

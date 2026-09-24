begin;

create table log_default_tags (
  id uuid primary key default uuid_generate_v4(),
  name varchar(50) not null unique
);

insert into log_default_tags
  (name)
values
  ('anime'),
  ('audiobook'),
  ('book'),
  ('chat'),
  ('chorusing'),
  ('comic'),
  ('conversation'),
  ('drama'),
  ('ebook'),
  ('fiction'),
  ('game'),
  ('grammar'),
  ('lyric'),
  ('news'),
  ('non-fiction'),
  ('online video'),
  ('podcast'),
  ('presentation'),
  ('shadowing'),
  ('social media'),
  ('srs'),
  ('textbook'),
  ('tv'),
  ('vocabulary'),
  ('web page');

commit;

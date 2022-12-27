begin;

create extension if not exists "uuid-ossp";

create table languages (
  code varchar(10) primary key not null,
  name varchar(100) not null
);

comment on column languages.code is 'See https://en.wikipedia.org/wiki/Wikipedia:WikiProject_Languages/List_of_ISO_639-3_language_codes_(2019)';

create table log_activity_types (
  id smallint primary key not null,
  name varchar(100) not null
);

create table log_units (
  id uuid primary key default uuid_generate_v4(),
  log_activity_type_id smallint not null,
  unit varchar(50) not null,
  modifier real not null,
  language_code varchar(10) -- could be null to indicate it's the fallback option for that language
);

create index log_units_log_activity_type_id on log_units(log_activity_type_id);

create table log_tags (
  id uuid primary key default uuid_generate_v4(),
  log_activity_type_id smallint not null,
  name varchar(50) not null
);

create index log_tags_log_activity_type_id on log_tags(log_activity_type_id);

create table logs (
  id uuid primary key default uuid_generate_v4(),
  contest_id uuid,
  user_id uuid not null,

  -- meta
  language_code varchar(10) not null,
  log_activity_type_id smallint not null,
  unit_id uuid not null,
  tags varchar(50)[] not null,

  -- scoring related
  amount real not null,
  modifier real not null,
  score real not null generated always as (amount * modifier) stored,

  -- optimize leaderboard fetching
  eligible_official_leaderboard boolean not null,
  "year" smallint not null generated always as (extract(year from created_at)) stored,

  created_at timestamp not null default now(),
  updated_at timestamp not null default now(),
  deleted_at timestamp default null
);

create index logs_year on logs(year);
create index logs_user_id on logs(user_id);
create index logs_user_id_contest on logs(user_id, contest_id);
create index logs_group_id on logs(user_id, contest_id);

-- Languages
insert into languages
  (code, name)
values
  ('ara', 'Arabic'),
	('zho', 'Chinese'),
	('hrv', 'Croatian'),
	('ces', 'Czech'),
	('dan', 'Danish'),
	('nld', 'Dutch'),
	('eng', 'English'),
	('epo', 'Esperanto'),
	('fin', 'Finnish'),
	('fra', 'French'),
	('deu', 'German'),
	('ell', 'Greek'),
	('heb', 'Hebrew'),
	('hun', 'Hungary'),
	('gle', 'Irish'),
	('ita', 'Italian'),
	('jpn', 'Japanese'),
	('kor', 'Korean'),
	('lat', 'Latin'),
	('nos', 'Norwegian'),
	('pol', 'Polish'),
	('por', 'Portuguese'),
	('rus', 'Russian'),
	('spa', 'Spanish'),
	('swe', 'Swedish'),
	('tha', 'Thai'),
	('tgl', 'Tagalog'),
	('tur', 'Turkish');

-- Activity types
insert into log_activity_types
  (id, name)
values
  (1, 'Reading'),
  (2, 'Listening'),
  (3, 'Writing'),
  (4, 'Speaking'),
  (5, 'Study');

-- Units
insert into log_units
  (log_activity_type_id, unit, modifier, language_code)
values
  (1, 'Page', 1, null),
  (1, '2 Column page', 1.6, 'jpa'),
  (1, 'Comic page', 0.2, null),
  (1, 'Sentence', 0.05, null),
  (1, 'Character', 0.000833333, null),
  (1, 'Character', 0.0025, 'jpa'),
  (1, 'Character', 0.0025, 'zho'),
  (1, 'Character', 0.0025, 'kor'),
  (2, 'Minute', 0.5, null),
  (2, 'Minute (high density)', 0.7, null),
  (3, 'Page', 1, null),
  (3, 'Sentence', 0.05, null),
  (3, 'Character', 0.000833333, null),
  (3, 'Character', 0.0025, 'jpa'),
  (3, 'Character', 0.0025, 'zho'),
  (3, 'Character', 0.0025, 'kor'),
  (4, 'Minute', 0.5, null),
  (4, 'Minute (high density)', 0.7, null),
  (5, 'Minute', 0.5, null);

-- Tags
insert into log_tags
  (log_activity_type_id, name)
values
  (1, 'Book'),
  (1, 'Ebook'),
  (1, 'Fiction'),
  (1, 'Non-fiction'),
  (1, 'Web page'),
  (1, 'Lyric'),
  (2, 'Audiobook'),
  (2, 'Anime'),
  (2, 'Fiction'),
  (2, 'Non-fiction'),
  (2, 'News'),
  (2, 'TV'),
  (2, 'Drama'),
  (2, 'Online video'),
  (2, 'Podcast'),
  (3, 'Fiction'),
  (3, 'Non-fiction'),
  (3, 'Social media'),
  (3, 'Chat'),
  (4, 'Conversation'),
  (4, 'Presentation'),
  (4, 'Shadowing'),
  (4, 'Chorusing'),
  (5, 'Grammar'),
  (5, 'Vocabulary'),
  (5, 'SRS'),
  (5, 'Textbook');

commit;
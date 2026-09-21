-- Use the same deterministic unit identities as the configuration fixtures.
update log_units set id = md5(unit_key || ':' || coalesce(language_code, ''))::uuid
where id <> md5(unit_key || ':' || coalesce(language_code, ''))::uuid;

insert into users (id, display_name, created_at, updated_at) values
  ('11111111-1111-4111-8111-111111111111', 'Stored Reader', '2026-01-01', '2026-09-13'),
  ('44444444-4444-4444-8444-444444444444', 'Reader Two', '2026-01-01', '2026-01-01'),
  ('99999999-9999-4999-8999-999999999999', 'Owner', '2026-01-01', '2026-01-01');
insert into contests (id, owner_user_id, owner_user_display_name, "private", contest_start, contest_end, registration_end, title, language_code_allow_list, activity_type_id_allow_list, official, created_at, updated_at) values
  ('f1111111-1111-4111-8111-111111111111', '99999999-9999-4999-8999-999999999999', 'stale', false, '2025-12-01', '2026-09-30', '2026-08-31', 'Detach fixture', '{jpn,eng}', '{1}', true, '2026-01-01', '2026-01-01');
insert into contest_registrations (id, contest_id, user_id, language_codes, created_at, updated_at) values
  ('e1111111-1111-4111-8111-111111111111', 'f1111111-1111-4111-8111-111111111111', '11111111-1111-4111-8111-111111111111', '{jpn,eng}', '2026-01-01', '2026-01-01');
insert into logs (id, user_id, language_code, log_activity_id, duration_seconds, computed_score, eligible_official_leaderboard, created_at, updated_at) values
  ('d1111111-1111-4111-8111-111111111111', '11111111-1111-4111-8111-111111111111', 'jpn', 1, 60, 1, true, '2026-09-01', '2026-09-01'),
  ('d2222222-2222-4222-8222-222222222222', '11111111-1111-4111-8111-111111111111', 'eng', 1, 60, 1, true, '2026-09-01', '2026-09-01'),
  ('d3333333-3333-4333-8333-333333333333', '44444444-4444-4444-8444-444444444444', 'eng', 1, 60, 1, true, '2026-09-01', '2026-09-01');
insert into contest_logs (contest_id, log_id, duration_seconds, computed_score) values
  ('f1111111-1111-4111-8111-111111111111', 'd1111111-1111-4111-8111-111111111111', 60, 1),
  ('f1111111-1111-4111-8111-111111111111', 'd2222222-2222-4222-8222-222222222222', 60, 1),
  ('f1111111-1111-4111-8111-111111111111', 'd3333333-3333-4333-8333-333333333333', 60, 1);
update logs set created_at='2026-09-01 01:00:00', updated_at='2026-09-02', amount=10, modifier=2, computed_score=0, unit_key='reading_page', unit_id=(select id from log_units where unit_key='reading_page' limit 1), description='A detailed log' where id='d1111111-1111-4111-8111-111111111111';
update logs set created_at='2026-09-03', computed_score=null where id='d2222222-2222-4222-8222-222222222222';
update logs set created_at='2026-09-04' where id='d3333333-3333-4333-8333-333333333333';
update contest_logs set amount=3, modifier=4, duration_seconds=120, computed_score=null where log_id='d1111111-1111-4111-8111-111111111111';
update contest_logs set computed_score=0 where log_id='d2222222-2222-4222-8222-222222222222';
insert into logs (id,user_id,language_code,log_activity_id,amount,modifier,duration_seconds,eligible_official_leaderboard,created_at,updated_at,deleted_at) values
 ('d4444444-4444-4444-8444-444444444444','11111111-1111-4111-8111-111111111111','eng',2,5,2,90,false,'2026-09-05','2026-09-05','2026-09-06'),
 ('d5555555-5555-4555-8555-555555555555','11111111-1111-4111-8111-111111111111','jpn',3,7,0.5,30,false,'2026-09-02','2026-09-02',null);
insert into contest_logs (contest_id,log_id,duration_seconds,computed_score) values
 ('f1111111-1111-4111-8111-111111111111','d4444444-4444-4444-8444-444444444444',90,50),
 ('f1111111-1111-4111-8111-111111111111','d5555555-5555-4555-8555-555555555555',30,8);
insert into log_tags (log_id,user_id,tag) values
 ('d1111111-1111-4111-8111-111111111111','11111111-1111-4111-8111-111111111111','zeta'), ('d1111111-1111-4111-8111-111111111111','11111111-1111-4111-8111-111111111111','alpha'), ('d1111111-1111-4111-8111-111111111111','11111111-1111-4111-8111-111111111111','comma, tag');
insert into contests (id,owner_user_id,owner_user_display_name,private,contest_start,contest_end,registration_end,title,language_code_allow_list,activity_type_id_allow_list,official,created_at,updated_at) values
 ('f2222222-2222-4222-8222-222222222222','99999999-9999-4999-8999-999999999999','stale',true,'2026-09-01','2026-09-30','2026-09-30','Private reading','{jpn}','{1}',false,'2026-01-01','2026-01-01');
insert into contest_registrations (id,contest_id,user_id,language_codes) values
 ('e2222222-2222-4222-8222-222222222222','f2222222-2222-4222-8222-222222222222','11111111-1111-4111-8111-111111111111','{jpn}');
insert into contest_logs (contest_id,log_id,duration_seconds,computed_score) values
 ('f2222222-2222-4222-8222-222222222222','d1111111-1111-4111-8111-111111111111',60,0);
update contests set activity_type_id_allow_list='{99}';
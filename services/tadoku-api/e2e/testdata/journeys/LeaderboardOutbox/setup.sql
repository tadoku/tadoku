delete from platform_scoring_config;
delete from scoring_rules;
delete from scoring_rule_sets;
insert into scoring_rule_sets (id, scope, version, status, created_at, published_at) values
 ('aaaaaaaa-aaaa-4aaa-8aaa-000000000001','platform',1,'published','2026-01-01','2026-01-02'),
 ('aaaaaaaa-aaaa-4aaa-8aaa-000000000002','platform',2,'draft','2026-02-01',null),
 ('aaaaaaaa-aaaa-4aaa-8aaa-000000000003','platform',3,'published','2026-02-02','2026-02-03');
insert into scoring_rules (id,rule_set_id,priority,stackable,activity_id,unit_key,language_code,tag,score_source,rate) values
 ('bbbbbbbb-bbbb-4bbb-8bbb-000000000002','aaaaaaaa-aaaa-4aaa-8aaa-000000000001',20,false,1,'reading_page',null,null,'amount',2),
 ('bbbbbbbb-bbbb-4bbb-8bbb-000000000001','aaaaaaaa-aaaa-4aaa-8aaa-000000000001',10,true,1,'reading_page',null,'manga','amount',0.5),
 ('bbbbbbbb-bbbb-4bbb-8bbb-000000000003','aaaaaaaa-aaaa-4aaa-8aaa-000000000001',30,false,5,null,null,null,'duration_minutes',0.5),
 ('bbbbbbbb-bbbb-4bbb-8bbb-000000000004','aaaaaaaa-aaaa-4aaa-8aaa-000000000002',5,false,2,null,null,null,'amount',3),
 ('bbbbbbbb-bbbb-4bbb-8bbb-000000000005','aaaaaaaa-aaaa-4aaa-8aaa-000000000003',5,false,5,null,null,null,'duration_minutes',0.8);
insert into platform_scoring_config (singleton,active_rule_set_id) values (true,'aaaaaaaa-aaaa-4aaa-8aaa-000000000001');
insert into users (id,display_name,created_at,updated_at) values
 ('11111111-1111-4111-8111-111111111111','Reader One','2026-01-01','2026-01-01'),
 ('22222222-2222-4222-8222-222222222222','Other Owner','2026-01-01','2026-01-01');
insert into contests (id,owner_user_id,owner_user_display_name,"private",contest_start,contest_end,registration_end,title,language_code_allow_list,activity_type_id_allow_list,official,created_at,updated_at) values
 ('f0000000-0000-4000-8000-000000000001','11111111-1111-4111-8111-111111111111','Reader One',false,'2026-09-01','2026-09-30','2026-09-20','Owned','{eng}','{1,5}',false,'2026-01-01','2026-01-01'),
 ('f0000000-0000-4000-8000-000000000002','22222222-2222-4222-8222-222222222222','Other Owner',false,'2026-09-01','2026-09-30','2026-09-20','Other','{eng}','{1}',false,'2026-01-01','2026-01-01'),
 ('f0000000-0000-4000-8000-000000000003','11111111-1111-4111-8111-111111111111','Reader One',false,'2026-09-01','2026-09-30','2026-09-20','Replace','{eng}','{5}',false,'2026-01-01','2026-01-01'),
 ('f0000000-0000-4000-8000-000000000004','11111111-1111-4111-8111-111111111111','Reader One',false,'2026-09-01','2026-09-30','2026-09-20','Inherit','{eng}','{1}',false,'2026-01-01','2026-01-01'),
 ('f0000000-0000-4000-8000-000000000005','11111111-1111-4111-8111-111111111111','Reader One',false,'2026-08-01','2026-09-10','2026-08-01','Ended','{eng}','{1}',false,'2026-01-01','2026-01-01');
insert into scoring_rule_sets (id,scope,contest_id,version,status,mode,fallback_rule_set_id,created_at,published_at) values
 ('cccccccc-cccc-4ccc-8ccc-000000000001','contest','f0000000-0000-4000-8000-000000000001',1,'published','override','aaaaaaaa-aaaa-4aaa-8aaa-000000000003','2026-03-01','2026-03-02'),
 ('cccccccc-cccc-4ccc-8ccc-000000000002','contest','f0000000-0000-4000-8000-000000000001',2,'draft','replace',null,'2026-04-01',null),
 ('cccccccc-cccc-4ccc-8ccc-000000000003','contest','f0000000-0000-4000-8000-000000000002',1,'published','replace',null,'2026-05-01','2026-05-02'),
 ('cccccccc-cccc-4ccc-8ccc-000000000004','contest','f0000000-0000-4000-8000-000000000003',1,'published','replace',null,'2026-06-01','2026-06-02');
insert into scoring_rules (id,rule_set_id,priority,stackable,activity_id,score_source,rate) values
 ('dddddddd-dddd-4ddd-8ddd-000000000001','cccccccc-cccc-4ccc-8ccc-000000000001',20,false,1,'amount',4),
 ('dddddddd-dddd-4ddd-8ddd-000000000002','cccccccc-cccc-4ccc-8ccc-000000000002',10,false,1,'amount',5),
 ('dddddddd-dddd-4ddd-8ddd-000000000003','cccccccc-cccc-4ccc-8ccc-000000000003',10,false,1,'amount',6);
update contests set scoring_rule_set_id='cccccccc-cccc-4ccc-8ccc-000000000001' where id='f0000000-0000-4000-8000-000000000001';
update contests set scoring_rule_set_id='cccccccc-cccc-4ccc-8ccc-000000000003' where id='f0000000-0000-4000-8000-000000000002';
update contests set scoring_rule_set_id='cccccccc-cccc-4ccc-8ccc-000000000004' where id='f0000000-0000-4000-8000-000000000003';
insert into contest_registrations (id,contest_id,user_id,language_codes) values
 ('eeeeeeee-eeee-4eee-8eee-000000000001','f0000000-0000-4000-8000-000000000001','11111111-1111-4111-8111-111111111111','{eng}'),
 ('eeeeeeee-eeee-4eee-8eee-000000000002','f0000000-0000-4000-8000-000000000002','22222222-2222-4222-8222-222222222222','{eng}'),
 ('eeeeeeee-eeee-4eee-8eee-000000000003','f0000000-0000-4000-8000-000000000003','11111111-1111-4111-8111-111111111111','{eng}'),
 ('eeeeeeee-eeee-4eee-8eee-000000000004','f0000000-0000-4000-8000-000000000004','11111111-1111-4111-8111-111111111111','{eng}'),
 ('eeeeeeee-eeee-4eee-8eee-000000000005','f0000000-0000-4000-8000-000000000005','11111111-1111-4111-8111-111111111111','{eng}');

update log_units
set id = '026e644b-ec51-61cd-a392-55231a6ab4a0'
where unit_key = 'reading_page' and language_code is null;

update contests set contest_end='2027-01-01', official=true where id='f0000000-0000-4000-8000-000000000001';
update contests set contest_end='2027-01-31' where id='f0000000-0000-4000-8000-000000000004';

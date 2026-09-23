delete from platform_scoring_config;
delete from scoring_rules;
delete from scoring_rule_sets;
insert into scoring_rule_sets (id, scope, version, status, created_at, published_at) values
 ('aaaaaaaa-aaaa-4aaa-8aaa-000000000001','platform',1,'published','2026-01-01','2026-01-02'),
 ('aaaaaaaa-aaaa-4aaa-8aaa-000000000002','platform',2,'draft','2026-02-01',null);
insert into scoring_rules (id,rule_set_id,priority,stackable,activity_id,unit_key,language_code,tag,score_source,rate) values
 ('bbbbbbbb-bbbb-4bbb-8bbb-000000000002','aaaaaaaa-aaaa-4aaa-8aaa-000000000001',20,false,1,'reading_page',null,null,'amount',2),
 ('bbbbbbbb-bbbb-4bbb-8bbb-000000000001','aaaaaaaa-aaaa-4aaa-8aaa-000000000001',10,true,1,'reading_page',null,'manga','amount',0.5),
 ('bbbbbbbb-bbbb-4bbb-8bbb-000000000003','aaaaaaaa-aaaa-4aaa-8aaa-000000000001',30,false,5,null,null,null,'duration_minutes',0.5),
 ('bbbbbbbb-bbbb-4bbb-8bbb-000000000004','aaaaaaaa-aaaa-4aaa-8aaa-000000000002',5,false,2,null,null,null,'amount',3);
insert into platform_scoring_config (singleton,active_rule_set_id) values (true,'aaaaaaaa-aaaa-4aaa-8aaa-000000000001');

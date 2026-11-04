delete from platform_scoring_config;
delete from scoring_rules;
delete from scoring_rule_sets;
insert into scoring_rule_sets (id, scope, version, status, created_at, published_at) values
 ('aaaaaaaa-aaaa-4aaa-8aaa-000000000005','platform',5,'published','2026-01-01','2026-01-02');
insert into scoring_rules (id,rule_set_id,priority,stackable,activity_id,unit_key,language_code,tag,score_source,rate) values
 ('bbbbbbbb-bbbb-4bbb-8bbb-000000000070','aaaaaaaa-aaaa-4aaa-8aaa-000000000005',70,false,1,'reading_page',null,'comic','amount',0.2),
 ('bbbbbbbb-bbbb-4bbb-8bbb-000000000075','aaaaaaaa-aaaa-4aaa-8aaa-000000000005',75,false,1,'reading_page',null,'manga','amount',0.2),
 ('bbbbbbbb-bbbb-4bbb-8bbb-000000000080','aaaaaaaa-aaaa-4aaa-8aaa-000000000005',80,false,1,'reading_page','jpn','two_column','amount',1.6),
 ('bbbbbbbb-bbbb-4bbb-8bbb-000000000100','aaaaaaaa-aaaa-4aaa-8aaa-000000000005',100,false,1,'reading_two_column_page','jpn',null,'amount',1.6),
 ('bbbbbbbb-bbbb-4bbb-8bbb-000000000110','aaaaaaaa-aaaa-4aaa-8aaa-000000000005',110,false,1,'reading_page',null,null,'amount',1),
 ('bbbbbbbb-bbbb-4bbb-8bbb-000000000120','aaaaaaaa-aaaa-4aaa-8aaa-000000000005',120,false,1,'reading_comic_page',null,null,'amount',0.2),
 ('bbbbbbbb-bbbb-4bbb-8bbb-000000000200','aaaaaaaa-aaaa-4aaa-8aaa-000000000005',200,false,2,null,null,null,'amount',0.4),
 ('bbbbbbbb-bbbb-4bbb-8bbb-000000000210','aaaaaaaa-aaaa-4aaa-8aaa-000000000005',210,true,2,'listening_dense_minutes',null,null,'amount',1.5),
 ('bbbbbbbb-bbbb-4bbb-8bbb-000000000220','aaaaaaaa-aaaa-4aaa-8aaa-000000000005',220,true,2,null,null,'dense','duration_minutes',1.5),
 ('bbbbbbbb-bbbb-4bbb-8bbb-000000000610','aaaaaaaa-aaaa-4aaa-8aaa-000000000005',610,false,2,null,null,null,'duration_minutes',0.4);
insert into platform_scoring_config (singleton,active_rule_set_id) values (true,'aaaaaaaa-aaaa-4aaa-8aaa-000000000005');

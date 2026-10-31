insert into users (id,display_name,created_at,updated_at) values
 ('11111111-1111-4111-8111-111111111111','Reader One','2026-01-01','2026-01-01'),
 ('22222222-2222-4222-8222-222222222222','Admin One','2026-01-01','2026-01-01');
insert into contests (id,owner_user_id,owner_user_display_name,"private",contest_start,contest_end,registration_end,title,language_code_allow_list,activity_type_id_allow_list,official,created_at,updated_at) values
 ('f0000000-0000-4000-8000-000000000001','22222222-2222-4222-8222-222222222222','Admin One',false,'2026-09-01','2026-09-30','2026-09-20','Official',null,'{1}',true,'2026-01-01','2026-01-01'),
 ('f0000000-0000-4000-8000-000000000002','11111111-1111-4111-8111-111111111111','Reader One',false,'2026-09-01','2026-09-30','2026-09-20','Unofficial','{eng}','{1}',false,'2026-01-01','2026-01-01');
insert into contest_registrations (id,contest_id,user_id,language_codes) values
 ('eeeeeeee-eeee-4eee-8eee-000000000001','f0000000-0000-4000-8000-000000000001','11111111-1111-4111-8111-111111111111','{eng}'),
 ('eeeeeeee-eeee-4eee-8eee-000000000002','f0000000-0000-4000-8000-000000000002','11111111-1111-4111-8111-111111111111','{eng}');

update log_units
set id = '026e644b-ec51-61cd-a392-55231a6ab4a0'
where unit_key = 'reading_page' and language_code is null;

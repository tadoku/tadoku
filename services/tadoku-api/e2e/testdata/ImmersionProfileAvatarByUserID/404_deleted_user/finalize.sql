update accounts
set email = 'deleted@example.test'
where user_id = '11111111-1111-4111-8111-111111111111';

update users
set deleted_at = '2026-09-01 00:00:00Z'
where id = '11111111-1111-4111-8111-111111111111';

delete from accounts where tenant = 'tadoku/prod' and user_id = '44444444-4444-4444-8444-444444444444';
update users set display_name = 'Deleted participant', deletion_locked_at = '2026-09-11 12:00:00', deleted_at = '2026-09-11 12:00:05'
where tenant = 'tadoku/prod' and id = '44444444-4444-4444-8444-444444444444';

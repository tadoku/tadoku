insert into tenants (key, kind)
values
  ('tadoku/identity-a-0123abcd', 'test'),
  ('tadoku/identity-b-0123abcd', 'test')
on conflict (key) do nothing;

insert into users (tenant, id, display_name, created_at, updated_at)
values
  ('tadoku/prod', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', 'Production Reader', '2026-01-01', '2026-01-01'),
  ('tadoku/identity-a-0123abcd', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb', 'Branch A Reader', '2026-01-02', '2026-01-02'),
  ('tadoku/identity-b-0123abcd', 'cccccccc-cccc-4ccc-8ccc-cccccccccccc', 'Branch B Reader', '2026-01-03', '2026-01-03');

insert into accounts (tenant, user_id, email, email_normalized, password_hash, created_at, updated_at)
select tenant, id, 'reader@example.test', 'reader@example.test',
  '$2b$12$TP8U/FQ8vCC/fzTCu5Fgve4WAHMgvG1jZ7V7bKWjDJRA.KeeC1xqK', created_at, updated_at
from users
where id in ('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb',
  'cccccccc-cccc-4ccc-8ccc-cccccccccccc');

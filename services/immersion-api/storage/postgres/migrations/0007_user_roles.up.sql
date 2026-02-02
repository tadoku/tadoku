begin;

create table user_roles (
  user_id uuid primary key references users(id),
  role varchar(50) not null check (role in ('admin', 'banned')),
  updated_at timestamp not null default now()
);

create index user_roles_role on user_roles(role);

commit;

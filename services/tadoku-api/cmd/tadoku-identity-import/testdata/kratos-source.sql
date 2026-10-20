-- Import-facing PostgreSQL tables from ory/kratos v26.2.0 migrations.
-- See the source-schema boundary in docs/docs/tadoku-api/identity-import.md.
create table networks (
    id uuid primary key,
    created_at timestamp not null,
    updated_at timestamp not null
);

create table identities (
    id uuid primary key,
    schema_id varchar(2048) not null,
    traits jsonb not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    nid uuid references networks(id) on update restrict on delete cascade,
    state varchar(255) not null default 'active',
    state_changed_at timestamp,
    metadata_public jsonb,
    metadata_admin jsonb,
    available_aal varchar(4),
    organization_id uuid,
    external_id varchar(64) check (external_id is null or external_id != '')
);

create table identity_credential_types (
    id uuid primary key,
    name varchar(32) not null
);

create table identity_credentials (
    id uuid primary key,
    config jsonb not null,
    identity_credential_type_id uuid not null references identity_credential_types(id) on delete cascade,
    identity_id uuid not null references identities(id) on delete cascade,
    created_at timestamp not null,
    updated_at timestamp not null,
    nid uuid references networks(id) on update restrict on delete cascade,
    version int not null default 0
);

insert into identity_credential_types (id, name) values
    ('78c1b41d-8341-4507-aa60-aff1d4369670', 'password'),
    ('6fa5e2e0-bfce-4631-b62b-cf2b0252b289', 'oidc'),
    ('5e29b036-aa47-457f-9fe6-aa8b854a752b', 'totp'),
    ('567a0730-7f48-4dd7-a13d-df87a51c245f', 'lookup_secret'),
    ('6b213fa0-e6ad-46cb-8878-b088d2ce2e3c', 'webauthn'),
    ('14f3b7e2-8725-4068-be39-8a796485fd97', 'code'),
    ('8d0ca544-9bf6-45d3-bd75-0bbb3aeba3c7', 'passkey'),
    ('7bddcf6c-f50e-4a18-9b0f-429114c33419', 'saml');

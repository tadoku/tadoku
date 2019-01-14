CREATE SEQUENCE user_seq;

CREATE TABLE users (
  id bigint check (id > 0) NOT NULL DEFAULT NEXTVAL ('user_seq'),
  email varchar(255) NOT NULL UNIQUE,
  display_name varchar(255) NOT NULL,
  password bytea NOT NULL,
  role smallint NOT NULL DEFAULT 0,
  preferences jsonb NOT NULL DEFAULT '{}'::jsonb,
  PRIMARY KEY (id)
);

ALTER SEQUENCE user_seq RESTART WITH 1;
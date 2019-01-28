CREATE SEQUENCE contest_seq;
CREATE SEQUENCE medium_seq;
CREATE SEQUENCE log_seq;

CREATE TABLE contests (
  id bigint check (id > 0) NOT NULL DEFAULT NEXTVAL ('contest_seq'),
  start date NOT NULL,
  "end" date NOT NULL,
  open boolean NOT NULL DEFAULT FALSE,
  PRIMARY KEY (id)
);

CREATE TABLE mediums (
  id smallint check (id > 0) NOT NULL DEFAULT NEXTVAL ('medium_seq'),
  description text NOT NULL,
  points float(2) NOT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE languages (
  iso_code varchar(3) NOT NULL,
  name varchar(50) NOT NULL,
  PRIMARY KEY (iso_code)
);

CREATE TABLE logs (
  id bigint check (id > 0) NOT NULL DEFAULT NEXTVAL ('log_seq'),
  contest_id bigint NOT NULL,
  user_id bigint NOT NULL,
  language_code varchar(3) NOT NULL,
  medium_id smallint NOT NULL,
  amount float(3) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL,
  deleted_at timestamp DEFAULT NULL,
  PRIMARY KEY (id)
);

CREATE INDEX logs_contest_id ON logs(contest_id);
CREATE INDEX logs_user_id ON logs(user_id);
CREATE INDEX logs_language_code ON logs(language_code);
CREATE INDEX logs_medium_id ON logs(medium_id);

ALTER SEQUENCE contest_seq RESTART WITH 1;
ALTER SEQUENCE medium_seq RESTART WITH 1;
ALTER SEQUENCE log_seq RESTART WITH 1;

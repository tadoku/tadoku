CREATE SEQUENCE contest_seq;
CREATE SEQUENCE medium_seq;

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

ALTER SEQUENCE contest_seq RESTART WITH 1;
ALTER SEQUENCE medium_seq RESTART WITH 1;

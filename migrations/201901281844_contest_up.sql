CREATE SEQUENCE contest_seq;

CREATE TABLE contests (
  id bigint check (id > 0) NOT NULL DEFAULT NEXTVAL ('contest_seq'),
  start date NOT NULL,
  "end" date NOT NULL,
  open boolean NOT NULL DEFAULT FALSE,
  PRIMARY KEY (id)
);

ALTER SEQUENCE contest_seq RESTART WITH 1;

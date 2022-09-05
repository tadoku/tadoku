alter table contest_logs add column "weighted_score" real default 0 not null;

-- backfill the data
update contest_logs set weighted_score = amount * 1 where medium_id = 1;
update contest_logs set weighted_score = amount * 0.2 where medium_id = 2;
update contest_logs set weighted_score = amount * 1 where medium_id = 3;
update contest_logs set weighted_score = amount * 0.1667 where medium_id = 4;
update contest_logs set weighted_score = amount * 0.05 where medium_id = 5;
update contest_logs set weighted_score = amount * 1 where medium_id = 6;
update contest_logs set weighted_score = amount * 1 where medium_id = 7;
update contest_logs set weighted_score = amount * 0.05 where medium_id = 8;

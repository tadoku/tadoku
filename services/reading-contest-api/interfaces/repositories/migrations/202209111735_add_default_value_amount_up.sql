-- adds a default value to amount column so we can deprecate it
alter table rankings alter column "amount" set default 0;

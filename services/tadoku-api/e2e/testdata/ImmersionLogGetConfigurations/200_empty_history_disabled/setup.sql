-- Preserve migration-seeded unit fields while making their random IDs repeatable.
update log_units
set id = md5(unit_key || ':' || coalesce(language_code, ''))::uuid
where id <> md5(unit_key || ':' || coalesce(language_code, ''))::uuid;

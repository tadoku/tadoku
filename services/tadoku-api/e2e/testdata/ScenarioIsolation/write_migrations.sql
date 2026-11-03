-- Test safety: deliberately unscoped; the seeding role must refuse migration metadata writes.
delete from schema_migrations;

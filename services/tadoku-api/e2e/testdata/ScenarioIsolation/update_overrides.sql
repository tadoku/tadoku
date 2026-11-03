-- Test safety: deliberately unscoped; the seeding role must refuse override writes.
update tenant_overrides set component = 'tampered';

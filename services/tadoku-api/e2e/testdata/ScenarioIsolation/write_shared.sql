-- Test safety: deliberately unscoped; the seeding role must refuse shared reference writes.
update languages set name = name || ' (tampered)';

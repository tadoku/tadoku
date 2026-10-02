-- Run outside a transaction so writes can continue while the index drops.
drop index concurrently scoring_rule_sets_tenant_contest_version;

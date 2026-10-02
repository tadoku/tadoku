-- Run outside a transaction so writes can continue while the index drops.
drop index concurrently scoring_rules_tenant_rule_set_priority;

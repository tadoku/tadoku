-- Run outside a transaction so writes can continue while the index drops.
drop index concurrently platform_scoring_config_tenant_singleton;

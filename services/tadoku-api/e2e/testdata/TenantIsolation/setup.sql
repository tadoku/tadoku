insert into tenants (key, kind) values ('e2e/isolation-0123abcd', 'test');
insert into scoring_rule_sets (tenant, id, scope, version, status, created_at, published_at)
values ('e2e/isolation-0123abcd', 'a0000000-0000-4000-8000-000000000001', 'platform', 1, 'published', '2026-01-01', '2026-01-02');
insert into scoring_rules (tenant, id, rule_set_id, priority, stackable, activity_id, unit_key, score_source, rate)
values ('e2e/isolation-0123abcd', 'b0000000-0000-4000-8000-000000000001', 'a0000000-0000-4000-8000-000000000001', 1, false, 1, 'reading_page', 'amount', 2);
insert into platform_scoring_config (tenant, singleton, active_rule_set_id)
values ('e2e/isolation-0123abcd', true, 'a0000000-0000-4000-8000-000000000001');

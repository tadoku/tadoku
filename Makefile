.PHONY: dev-up dev-down dev-reset dev-seed dev-logs

dev-up:
	dev up --task migrate --task seed

dev-down:
	dev down

dev-reset:
	@echo 'dev-reset is disabled: the legacy reset targets the old shared stack. Use dev down for overlay cleanup; database deletion requires an explicitly approved, scoped runbook.' >&2
	@exit 1

dev-seed:
	./scripts/dev/seed-db.sh

dev-logs:
	dev logs

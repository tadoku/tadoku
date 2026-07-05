.PHONY: dev-up dev-down dev-reset dev-seed dev-logs

dev-up:
	tilt up

dev-down:
	tilt down

dev-reset:
	./scripts/dev/reset-env.sh

dev-seed:
	./scripts/dev/seed-db.sh

dev-logs:
	tilt logs

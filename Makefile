.PHONY: dev-up dev-down dev-reset dev-seed dev-logs

dev-up:
	dev up --task migrate --task seed

dev-down:
	dev down

dev-reset:
	./scripts/dev/reset-env.sh

dev-seed:
	./scripts/dev/seed-db.sh

dev-logs:
	dev logs

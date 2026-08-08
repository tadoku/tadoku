.PHONY: dev-up dev-down dev-reset dev-seed dev-verify-scoring dev-logs

dev-up:
	tilt up

dev-down:
	tilt down

dev-reset:
	./scripts/dev/reset-env.sh

dev-seed:
	./scripts/dev/seed-db.sh

dev-verify-scoring:
	./scripts/dev/verify-scoring.sh

dev-logs:
	tilt logs

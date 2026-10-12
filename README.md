# Tadoku Monorepo

Developer documentation lives in [`docs/docs`](docs/docs) and is published at https://tadoku.github.io/tadoku/.

## Dev Environment

Branches run on the shared `homelab-talos-dev` cluster with dev-cli. Installation,
branch tenants, seed accounts, routing and cleanup are in
[Development environment](docs/docs/develop/environment.md).
For agent-driven live edits, use the [dev-cli skill](.agents/skills/dev-cli/SKILL.md)
(`$dev-cli` where repository skill discovery is supported).

To verify a change in the browser, follow the
[verify-tadoku skill](.agents/skills/verify-tadoku/SKILL.md) (`$verify-tadoku`
where repository skill discovery is supported) and its
[feature map](.agents/skills/verify-tadoku/references/features/README.md).

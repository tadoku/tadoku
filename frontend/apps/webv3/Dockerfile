FROM docker.io/node:22-bookworm-slim@sha256:c3de60bf2f9dd0ac6370e6117950ff62d6e339527e7472301c9c78a017978392 AS build

WORKDIR /app
RUN corepack enable && corepack prepare pnpm@10.10.0 --activate
COPY pnpm-lock.yaml pnpm-workspace.yaml package.json .npmrc ./
RUN pnpm fetch
COPY apps/webv3 ./apps/webv3
COPY packages/paper-ui ./packages/paper-ui
RUN pnpm install --frozen-lockfile --offline --filter webv3...
ENV NODE_OPTIONS=--max-old-space-size=3072
RUN pnpm --filter paper-ui build && pnpm --filter webv3 build

FROM docker.io/node:22-bookworm-slim@sha256:c3de60bf2f9dd0ac6370e6117950ff62d6e339527e7472301c9c78a017978392 AS release
LABEL org.opencontainers.image.source="https://github.com/tadoku/tadoku"
ENV NODE_ENV=production HOST=0.0.0.0 PORT=3000 TADOKU_WEB_URL=https://tadoku.app TADOKU_API_URL=https://tadoku.app/api/internal
WORKDIR /app
COPY --from=build --chown=node:node /app/apps/webv3/.output ./
USER 1000:1000
EXPOSE 3000
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s CMD node -e "fetch('http://127.0.0.1:3000/healthz').then(r=>{if(!r.ok)process.exit(1)}).catch(()=>process.exit(1))"
CMD ["node", "server/index.mjs"]

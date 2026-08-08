FROM docker.io/node:20-alpine@sha256:fb4cd12c85ee03686f6af5362a0b0d56d50c58a04632e6c0fb8363f609372293 AS build

WORKDIR /app
RUN npm install -g pnpm@10.10.0 && pnpm config set store-dir .pnpm-store
COPY pnpm-lock.yaml .npmrc* ./
RUN pnpm fetch
COPY . .
RUN pnpm install --frozen-lockfile --offline --ignore-scripts --workspace-root --filter paper-styleguide...
RUN pnpm --filter paper-ui build && pnpm --filter paper-styleguide build

FROM nginxinc/nginx-unprivileged:1.27-alpine@sha256:65e3e85dbaed8ba248841d9d58a899b6197106c23cb0ff1a132b7bfe0547e4c0 AS release

COPY apps/paper-styleguide/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/apps/paper-styleguide/dist /usr/share/nginx/html

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=3s --retries=3 CMD wget -q -O /dev/null http://127.0.0.1:8080/healthz || exit 1

# Builder image
FROM docker.io/node:20-alpine AS build

ARG PROJECT_NAME

WORKDIR /app

# Set up pnpm
RUN npm install -g pnpm@10.10.0 && pnpm config set store-dir .pnpm-store
COPY pnpm-lock.yaml .npmrc* ./
RUN pnpm fetch

# Build
COPY . .
RUN pnpm install --frozen-lockfile --offline --ignore-scripts --workspace-root --filter ${PROJECT_NAME}...
RUN pnpm --filter=${PROJECT_NAME}... run build
# WORKAROUND FOR: https://github.com/vercel/next.js/discussions/39432
RUN pnpm install --prod --frozen-lockfile --offline --shamefully-hoist --ignore-scripts --workspace-root --filter ${PROJECT_NAME} && \
  cp -Lr ./node_modules ./node_modules_temp && \
  rm -rf ./node_modules_temp/.cache && \
  rm -rf ./node_modules_temp/.pnpm
# END WORKAROUND

# Runtime image
FROM docker.io/node:20-alpine AS release

ARG PROJECT_NAME

ENV PORT=3000

WORKDIR /app
RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

COPY --from=build --chown=nextjs:nodejs /app/apps/${PROJECT_NAME}/.next/standalone ./
COPY --from=build --chown=nextjs:nodejs /app/apps/${PROJECT_NAME}/.next/static* ./apps/${PROJECT_NAME}/.next/static
COPY --from=build --chown=nextjs:nodejs /app/apps/${PROJECT_NAME}/public* ./apps/${PROJECT_NAME}/public

RUN ln -s /app/apps/${PROJECT_NAME}/server.js /app/server.js

USER nextjs

EXPOSE 3000

CMD ["node", "server.js"]
# Install deps
FROM node:12 AS base
WORKDIR /base
COPY . .
RUN yarn install

# Build app
FROM base AS build
ENV NODE_ENV=production
WORKDIR /build
COPY --from=base /base ./
RUN yarn run build

# Create production container
FROM node:12 AS production
ENV NODE_ENV=production
WORKDIR /app
COPY --from=build /build/package.json /build/yarn.lock /build/server.js ./
# Need to add dummy RUN statements so Docker doesn't crash...
# Ref. https://github.com/moby/moby/issues/37965
RUN true
COPY --from=build /build/next.config.js /build/config.js ./
RUN true
COPY --from=build /build/yarn.lock ./
RUN true
COPY --from=build /build/.next ./.next
RUN true
COPY --from=build /build/public ./public
RUN yarn add next express http-proxy-middleware --frozen-lockfile --production && yarn cache clean

# Running the app
EXPOSE 3000
CMD [ "yarn", "start" ]

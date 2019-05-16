FROM node:10

WORKDIR /usr/src/app


COPY . .

RUN yarn install

# Building app
RUN yarn run build

# Running the app
CMD [ "yarn", "start" ]
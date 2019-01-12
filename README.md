# Tadoku API

[![CircleCI](https://circleci.com/gh/tadoku/api/tree/master.svg?style=svg)](https://circleci.com/gh/tadoku/api/tree/master)

## Getting started

### 1. Setup environment

- Install [`direnv`](https://direnv.net/)
- Copy over the default environment: `$ cp .env{,.sample}`
- Go over the file and make sure the environment variables are correct for your env (eg. database url)
- Allow direnv `$ direnv allow`

## Commands

### Build project

```sh
$ make all
```

### Run tests

```sh
$ make test
```
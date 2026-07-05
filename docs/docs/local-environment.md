---
sidebar_position: 3
title: Development Environment
---

# Development Environment

Tadoku is made up of several services working together. It can be quite difficult to set up a local development environment with all the required services linked up together. This is a requirement for anyone to be productive in this project, and is also why we've provided a development environment for you.

We use [Tilt](https://tilt.dev/) to deploy all our backend services & dependencies to a Kubernetes cluster. Tilt can target either a local cluster (the `orbstack` context) or the shared `tadoku-dev` lab cluster; when targeting `tadoku-dev`, built images are pushed to the `registry.tadoku.lab` registry. The environment includes both the backend services and the frontend apps. Each frontend also has a development mode which is configured to connect to this environment.

## Getting Started

1. Install [Helm](https://helm.sh/docs/intro/install/).
2. Install [Bazel](https://docs.bazel.build/bazel-overview.html).
3. Install [Tilt](https://docs.tilt.dev/install.html).
4. Install [kubectl](https://kubernetes.io/docs/tasks/tools/).
5. Point kubectl at the cluster you want to use (see below).
6. Read the [Getting Started Tutorial](https://docs.tilt.dev/tutorial.html) for Tilt to get familiar with it.
7. Run `$ tilt up` in the root of this repository.
8. Some services will have a database seed script, these can be manually triggered from within Tilt when needed.
9. Access the environment (see below for the hostnames of each cluster).

### Option A: Local cluster (`orbstack`)

Run a local Kubernetes cluster (e.g. via [OrbStack](https://orbstack.dev/)) and make sure your kubectl context is named `orbstack`:

```sh
kubectl config use-context orbstack
```

Built images stay in your local docker daemon. Access the environment from `http://langlog.be`, a domain reserved to serve a local dev instance of Tadoku.

### Option B: Shared lab cluster (`tadoku-dev`)

Fetch the dev-cluster kubeconfig:

```sh
./infra/dev/kubeconfig.sh
export KUBECONFIG="$HOME/.kube/tadoku-dev.yaml"
kubectl get nodes
```

The script reads `/etc/rancher/k3s/k3s.yaml` from `io@ct200.lab`, rewrites the API server to `https://ct200.lab:6443`, sets the context to `tadoku-dev`, and stores the result as `~/.kube/tadoku-dev.yaml`. Override the SSH target with `TADOKU_DEV_K3S_SSH_TARGET` or pass a destination path as the first argument. The host (`TADOKU_DEV_K3S_HOST`), read command (`TADOKU_DEV_K3S_READ_CMD`), TLS server name (`TADOKU_DEV_K3S_TLS_SERVER_NAME`), and output path (`TADOKU_DEV_KUBECONFIG`) can also be overridden via environment variables.

Built images are pushed to `registry.tadoku.lab`. Access the dev cluster through the `.tadoku.lab` hostnames.

## Can't connect connect to service/database

It's possible that the containers for a particular service or database have been shut down due resource constraints. In this case you can restart the service manually from the Tilt dashboard. If a database is unreachable it might be useful to restart the Tilt cluster.

## Connecting to a database within Tilt

We have configured a [pgweb instance](https://github.com/sosedoff/pgweb) for access to our PostgreSQL databases. The connection details can be found from the Kubernetes secrets and/or manifest files.

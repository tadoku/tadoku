---
sidebar_position: 3
title: Development Environment
---

# Development Environment

Tadoku is made up of several services working together. It can be quite difficult to set up a local development environment with all the required services linked up together. This is a requirement for anyone to be productive in this project, and is also why we've provided a development environment for you.

We use [Tilt](https://tilt.dev/) to deploy all our backend services & dependencies to a Kubernetes cluster. Tilt can target either a local cluster or a shared development cluster; when targeting the shared cluster, built images are pushed to the registry configured in `tilt_config.json`. The environment includes both the backend services and the frontend apps. Each frontend also has a development mode which is configured to connect to this environment.

## Getting Started

1. Install [Helm](https://helm.sh/docs/intro/install/).
2. Install [Bazel](https://docs.bazel.build/bazel-overview.html).
3. Install [Tilt](https://docs.tilt.dev/install.html).
4. Install [kubectl](https://kubernetes.io/docs/tasks/tools/).
5. Point kubectl at the cluster you want to use (see below).
6. Copy `tilt_config.json.example` to `tilt_config.json` and set the local hostnames for your cluster. This file is gitignored.
7. Read the [Getting Started Tutorial](https://docs.tilt.dev/tutorial.html) for Tilt to get familiar with it.
8. Run `$ tilt up` in the root of this repository.
9. Some services will have a database seed script, these can be manually triggered from within Tilt when needed.
10. Access the environment using the hostnames from your local Tilt config.

### Option A: Local cluster

Run a local Kubernetes cluster and select its kubectl context. Tilt uses `orbstack` by default for local development; set `local_k8s_context` in `tilt_config.json` when your local context has another name (the `TADOKU_LOCAL_K8S_CONTEXT` environment variable overrides it):

```sh
kubectl config use-context docker-desktop
# and in tilt_config.json: "local_k8s_context": "docker-desktop"
```

Built images stay in your local docker daemon. Because backend images are built with Bazel and never pushed to a registry for local contexts, the cluster must be able to run images straight from the host Docker daemon (e.g. OrbStack or Docker Desktop). Clusters with their own container runtime, such as kind or minikube, would need an extra image-load step (e.g. `kind load`) that is not supported yet. Access the environment using the `local.hosts` values in `tilt_config.json`.

### Option B: Shared development cluster

Prerequisite: the configured registry hostname must resolve from your machine, and your Docker daemon must trust the registry endpoint (via an insecure-registry entry for HTTP, or the platform TLS certificate once available) before Tilt can push images.

Copy `tilt_config.json.example` to the gitignored `tilt_config.json`, then replace the placeholder values with your private operator values. Keep real hostnames, registry names, and context names in private config only.

Fetch the dev-cluster kubeconfig after creating `infra/dev/.env.local` from `infra/dev/.env.example`:

```sh
cp infra/dev/.env.example infra/dev/.env.local
# Edit infra/dev/.env.local with your private operator values.
./infra/dev/kubeconfig.sh
export KUBECONFIG="$HOME/.kube/<shared-context>.yaml"
kubectl get nodes
```

The script reads `/etc/rancher/k3s/k3s.yaml` from `TADOKU_DEV_K3S_SSH_TARGET`, rewrites the API server to `https://$TADOKU_DEV_K3S_HOST:6443`, sets the context to `TADOKU_DEV_K8S_CONTEXT`, and stores the result as `~/.kube/<shared-context>.yaml` unless `TADOKU_DEV_KUBECONFIG` or a destination argument is provided. The host (`TADOKU_DEV_K3S_HOST`), SSH target (`TADOKU_DEV_K3S_SSH_TARGET`), context (`TADOKU_DEV_K8S_CONTEXT`), read command (`TADOKU_DEV_K3S_READ_CMD`), TLS server name (`TADOKU_DEV_K3S_TLS_SERVER_NAME`), and output path (`TADOKU_DEV_KUBECONFIG`) can be set in `infra/dev/.env.local` or as environment variables; environment variables take precedence over `.env.local`.

Built images are pushed to `shared.registry` from `tilt_config.json`. The app and auth/admin hostnames come from the `shared.hosts` block.

## Can't connect connect to service/database

It's possible that the containers for a particular service or database have been shut down due resource constraints. In this case you can restart the service manually from the Tilt dashboard. If a database is unreachable it might be useful to restart the Tilt cluster.

## Connecting to a database within Tilt

We have configured a [pgweb instance](https://github.com/sosedoff/pgweb) for access to our PostgreSQL databases. The connection details can be found from the Kubernetes secrets and/or manifest files.

---
sidebar_position: 3
title: Development Environment
---

# Development Environment

Tadoku is made up of several services working together. It can be quite difficult to set up a local development environment with all the required services linked up together. This is a requirement for anyone to be productive in this project, and is also why we've provided a development environment for you.

We use [Tilt](https://tilt.dev/) to deploy all our backend services & dependencies to a Kubernetes cluster. Tilt can target either a local cluster or the shared `dev-lab` cluster; when targeting `dev-lab`, built images are pushed to the registry configured in `tilt_config.json`. The environment includes both the backend services and the frontend apps. Each frontend also has a development mode which is configured to connect to this environment.

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

Run a local Kubernetes cluster and select its kubectl context. Tilt uses `orbstack` by default for local development. For a persistent local override, set the context and cluster type in `tilt_config.json`:

```json
{
  "local_k8s_context": "kind-tadoku",
  "local_cluster_type": "kind",
  "local_cluster_name": "tadoku"
}
```

Environment variables still work for one-off runs and take precedence over the file:

```sh
kubectl config use-context docker-desktop
export TADOKU_LOCAL_K8S_CONTEXT=docker-desktop
export TADOKU_LOCAL_CLUSTER_TYPE=docker-desktop
```

For local contexts, backend images are built with Bazel and are not pushed to a registry. OrbStack and Docker Desktop use the daemon-sharing path, while kind and minikube run an image-load step after each backend image build (`kind load docker-image` or `minikube image load`). If your kind cluster or minikube profile name is not obvious from the kubectl context, set `local_cluster_name` in `tilt_config.json` or `TADOKU_LOCAL_CLUSTER_NAME`. The cluster type is inferred from the context name; accepted `local_cluster_type` values are `kind`, `minikube`, and the daemon-sharing types `orbstack`, `docker-desktop`, `shared-daemon`, and `none` (the last two skip the image-load step for any other daemon-sharing runtime). Access the environment using the `local.hosts` values in `tilt_config.json`.

### Option B: Shared lab cluster (`dev-lab`)

Prerequisite: the configured registry hostname must resolve from your machine, and your Docker daemon must trust the registry endpoint (via an insecure-registry entry for HTTP, or the lab TLS certificate once available) before Tilt can push images.

Fetch the dev-cluster kubeconfig after creating `.env.local` from `.env.local.example`:

```sh
cp .env.local.example .env.local
# Edit .env.local with the shared cluster host and SSH target.
./infra/dev/kubeconfig.sh
export KUBECONFIG="$HOME/.kube/dev-lab.yaml"
kubectl get nodes
```

The script reads `/etc/rancher/k3s/k3s.yaml` from `TADOKU_DEV_K3S_SSH_TARGET`, rewrites the API server to `https://${TADOKU_DEV_K3S_HOST}:6443`, sets the context to `dev-lab`, and stores the result as `~/.kube/dev-lab.yaml`. Override the SSH target with `TADOKU_DEV_K3S_SSH_TARGET` or pass a destination path as the first argument. The host (`TADOKU_DEV_K3S_HOST`), read command (`TADOKU_DEV_K3S_READ_CMD`), TLS server name (`TADOKU_DEV_K3S_TLS_SERVER_NAME`), and output path (`TADOKU_DEV_KUBECONFIG`) can also be overridden via environment variables.

Built images are pushed to `shared.registry` from `tilt_config.json`. The app and auth/admin hostnames come from the `shared.hosts` block.

## Can't connect connect to service/database

It's possible that the containers for a particular service or database have been shut down due resource constraints. In this case you can restart the service manually from the Tilt dashboard. If a database is unreachable it might be useful to restart the Tilt cluster.

## Connecting to a database within Tilt

We have configured a [pgweb instance](https://github.com/sosedoff/pgweb) for access to our PostgreSQL databases. The connection details can be found from the Kubernetes secrets and/or manifest files.

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
9. Seed data is applied by the `dev-seed` Tilt resource once the services are up (see below); it can also be re-run manually from within Tilt or via `make dev-seed`.
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

When `TADOKU_LOCAL_K8S_CONTEXT` is set, the `local_cluster_type` and `local_cluster_name` values from `tilt_config.json` are ignored (they describe the file's own context, not the override); set the matching `TADOKU_LOCAL_CLUSTER_TYPE`/`TADOKU_LOCAL_CLUSTER_NAME` env vars or rely on inference from the context name. If the type is inferred as `shared-daemon`, Tilt prints a warning since locally built images will not be loaded into the cluster.

For local contexts, backend images are built with Bazel and are not pushed to a registry. OrbStack and Docker Desktop use the daemon-sharing path, while kind and minikube run an image-load step after each backend image build (`kind load docker-image` or `minikube image load`). If your kind cluster or minikube profile name is not obvious from the kubectl context, set `local_cluster_name` in `tilt_config.json` or `TADOKU_LOCAL_CLUSTER_NAME`. The cluster type is inferred from the context name; accepted `local_cluster_type` values are `kind`, `minikube`, and the daemon-sharing types `orbstack`, `docker-desktop`, `shared-daemon`, and `none` (the last two skip the image-load step for any other daemon-sharing runtime). Access the environment using the `local.hosts` values in `tilt_config.json`.

### Option B: Shared development cluster

Prerequisite: the configured registry hostname must resolve from your machine, and your Docker daemon must trust the registry endpoint (via an insecure-registry entry for HTTP, or the platform TLS certificate once available) before Tilt can push images.

Copy `tilt_config.json.example` to the gitignored `tilt_config.json`, then replace the placeholder values with your private operator values. Keep real hostnames, registry names, and context names in private config only. The context keys can also be set via environment variables (`shared_k8s_context` → `TADOKU_SHARED_K8S_CONTEXT`, `local_k8s_context` → `TADOKU_LOCAL_K8S_CONTEXT`); environment variables take precedence over `tilt_config.json`. Shared-cluster mode stays disabled until a shared context is configured via `shared_k8s_context` or `TADOKU_SHARED_K8S_CONTEXT`; there is no committed default.

Fetch the dev-cluster kubeconfig after creating `infra/dev/.env.local` from `infra/dev/.env.example`:

```sh
cp infra/dev/.env.example infra/dev/.env.local
# Edit infra/dev/.env.local with your private operator values.
./infra/dev/kubeconfig.sh
export KUBECONFIG="$HOME/.kube/<shared-context>.yaml"
kubectl get nodes
```

The script reads `/etc/rancher/k3s/k3s.yaml` from `TADOKU_DEV_K3S_SSH_TARGET`, rewrites the API server to `https://$TADOKU_DEV_K3S_HOST:6443`, sets the context to `TADOKU_DEV_K8S_CONTEXT`, and stores the result as `~/.kube/<shared-context>.yaml` unless `TADOKU_DEV_KUBECONFIG` or a destination argument is provided. The host (`TADOKU_DEV_K3S_HOST`), SSH target (`TADOKU_DEV_K3S_SSH_TARGET`), and context (`TADOKU_DEV_K8S_CONTEXT`, falling back to `TADOKU_SHARED_K8S_CONTEXT`) are required and the script exits with an error if any is unset. The read command (`TADOKU_DEV_K3S_READ_CMD`), TLS server name (`TADOKU_DEV_K3S_TLS_SERVER_NAME`), and output path (`TADOKU_DEV_KUBECONFIG`) are optional. All of these can be set in `infra/dev/.env.local` or as environment variables; environment variables take precedence over `.env.local`. Set `TADOKU_DEV_KUBECONFIG_ENV` to read the env file from a different path.

Built images are pushed to `shared.registry` from `tilt_config.json`. The app and auth/admin hostnames come from the `shared.hosts` block.

### HTTPS and TLS

Each environment block in `tilt_config.json` controls how the dev stack is exposed:

- `scheme` (`http` or `https`): scheme used for the generated app/auth/admin URLs. Defaults to `https` for the shared environment and `http` for local. With `https`, session cookies are marked secure (`NEXT_PUBLIC_COOKIE_SECURE`) and Kratos runs with `development: false` unless overridden.
- `tls.enabled`: adds a `tls` block and a `cert-manager.io/cluster-issuer` annotation to the app, auth, and admin ingresses. Defaults to true on the shared environment when `scheme` is `https`, false otherwise. When enabled, `tls.cluster_issuer` is required; `tls.secret_names.{app,auth,admin}` override the default certificate secret names (`tadoku-dev-{app,auth,admin}-tls`).
- `ssl_redirect`: adds nginx annotations that redirect HTTP traffic to HTTPS on all dev ingresses. Defaults to the value of `tls.enabled`. Only supported with `ingress_class: "nginx"` — rendering fails otherwise, so set it to `false` on non-nginx clusters.
- `kratos_development`: toggles Kratos development mode. Defaults to true unless `scheme` is `https`.

See `tilt_config.json.example` for a complete example of both environments.

### Private infrastructure

Operators can include a private Tiltfile (kept outside this repository) by setting `TADOKU_PRIVATE_INFRA_TILTFILE` to its path before running `tilt up`; it is skipped when unset or when the file does not exist.

## Seeding and resetting the dev database

The root `Makefile` wraps the common dev-environment commands:

```sh
make dev-up      # start Tilt
make dev-down    # stop Tilt-managed resources
make dev-seed    # rerun idempotent seed data only (scripts/dev/seed-db.sh)
make dev-reset   # delete/recreate the dev DB, rerun migrations, and seed (scripts/dev/reset-env.sh)
make dev-logs    # stream Tilt logs
```

Seeding (`dev-seed`, also a Tilt resource that runs automatically once the backend services are ready) creates two Kratos identities — an admin `dev@tadoku.app` and a regular user `reader@tadoku.app`, both with password `tadoku` — grants the admin a Keto admin relation, and loads deterministic contests, activity logs, profile, and content data into the `immersion`, `profile`, and `content` databases. The seed is idempotent and safe to re-run: identities created by the seed carry a `seeded_by: tadoku-dev-seed` admin metadata marker, and re-running the seed refreshes their password to the currently configured value, so password overrides take effect on the next `dev-seed`. Identities without the marker (real or manually created accounts) are never modified, even if their email matches. Defaults can be overridden with `TADOKU_DEV_NAMESPACE`, `TADOKU_DEV_DB_PASSWORD`, `TADOKU_DEV_ADMIN_EMAIL`/`TADOKU_DEV_ADMIN_PASSWORD`, and `TADOKU_DEV_READER_EMAIL`/`TADOKU_DEV_READER_PASSWORD`.

Resetting (`dev-reset`, also available as a manual-only `dev-reset` Tilt resource) is destructive: it deletes the Zalando operator-managed `tadoku-dev-db` cluster and its persistent volume claims, reapplies the `postgresql` custom resource, restarts the backend services so their startup migrations run against the fresh database, and then reseeds. The script refuses to run unless the current kubectl context matches a known dev context (`shared_k8s_context`/`local_k8s_context` from `tilt_config.json`, or the `TADOKU_SHARED_K8S_CONTEXT`/`TADOKU_LOCAL_K8S_CONTEXT` env vars), and requires typing the context name to confirm when targeting the shared cluster. Ordinary `tilt down`/`tilt up` keeps data since the database uses persistent volumes.

## Private development registry

Tilt can push dev images to a private registry and create pull secrets for the namespaces that run Tilt-built images. Configure the pull-side credentials with local environment variables before running Tilt:

```sh
export TADOKU_DEV_REGISTRY_HOST="<registry-host>"
export TADOKU_DEV_REGISTRY_USERNAME="registry-pull"
export TADOKU_DEV_REGISTRY_PASSWORD="<registry-pull-password>"
```

The host value is also used as Tilt's `default_registry`, so authenticate Docker separately with the push account:

```sh
docker login "$TADOKU_DEV_REGISTRY_HOST" --username registry-push
```

Do not commit real registry hosts, usernames beyond the documented role names, passwords, or generated Secret YAML.

## Can't connect connect to service/database

It's possible that the containers for a particular service or database have been shut down due resource constraints. In this case you can restart the service manually from the Tilt dashboard. If a database is unreachable it might be useful to restart the Tilt cluster.

## Connecting to a database within Tilt

We have configured a [pgweb instance](https://github.com/sosedoff/pgweb) for access to our PostgreSQL databases. The connection details can be found from the Kubernetes secrets and/or manifest files.

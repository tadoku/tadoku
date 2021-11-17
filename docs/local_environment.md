# Local Development Environment

Tadoku is made up of several services working together. It can be quite difficult to set up a local development environment with all the required services linked up together. This is a requirement for anyone to be productive in this project, and is also why we've provided a development environment for you.

We use [Tilt](https://tilt.dev/) to spin up a local Kubernetes cluster with all our backend services & dependencies. We've decided to leave out the frontend packages from this environment for now, as it turned out to be quite resource intensive when doing so. Each frontend will have some sort of development mode included which is configured to connect to this environment.

## Getting Started

1. Install [Helm](https://helm.sh/docs/intro/install/).
2. Install [Bazel](https://docs.bazel.build/bazel-overview.html)
3. Install [Tilt](https://docs.tilt.dev/install.html).
4. Read the [Getting Started Tutorial](https://docs.tilt.dev/tutorial.html) for Tilt to get familiar with it.
5. Run `$ tilt up` in the root of this repository.
6. Some services will have a database seed script, these can be manually triggered from within Tilt when needed.

## Can't connect connect to service/database

It's possible that the containers for a particular service or database have been shut down due resource constraints. In this case you can restart the service manually from the Tilt dashboard. If a database is unreachable it might be useful to restart the Tilt cluster.

## Connecting to a database within Tilt

We have configured a [pgweb instance](https://github.com/sosedoff/pgweb) for access to our PostgreSQL databases. The connection details can be found from Kubernetes.

Given the following database configuration we can derive the connections as follows:

```yaml
apiVersion: "acid.zalan.do/v1"
kind: postgresql
metadata:
  name: reading-contest-api-db
  labels:
    app: reading-contest-api-db
spec:
  teamId: "tadoku"
  volume:
    size: 1Gi
  numberOfInstances: 1
  users:
    tadoku_user: []
  databases:
    tadoku: tadoku_user
  postgresql:
    version: "13"
```

```sh
# Host, derived from `metadata.name`
reading-contest-api-db

# Username, default superuser
postgres

# Password, secret name has the following pattern: `$username.$database_name.credentials.postgresql.acid.zalan.do`
$ kubectl get secret postgres.reading-contest-api-db.credentials.postgresql.acid.zalan.do -o json | jq -r .data.password | base64 --decode`

# Database name, derived from `spec.databases.$database_name`
tadoku
```

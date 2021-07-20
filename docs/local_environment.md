# Local Development Environment

Tadoku is made up of several services working together. It can be quite difficult to set up a local development environment with all the required services linked up together. This is a requirement for anyone to be productive in this project, and is also why we've provided a development environment for you.

We use [Tilt](https://tilt.dev/) to spin up a local Kubernetes cluster with all our backend services & dependencies. We've decided to leave out the frontends from this environment for now, as it turned out to be quite resource intensive when doing so. Each frontend will have some sort of development mode included which is configured to connect to this environment.

## Getting Started

1. Install Tilt according to their [installation instructions](https://docs.tilt.dev/install.html).
2. Read the [Getting Started Tutorial](https://docs.tilt.dev/tutorial.html) for Tilt to get familiar with it.
3. Run `$ tilt up` in the root of this repository.
4. Some services will have a database seed script, these can be manually triggered from within Tilt when needed.

## Connecting to a database within Tilt

We have configured a [pgweb instance](https://github.com/sosedoff/pgweb) for access to our PostgreSQL databases. The connection details can be found from Kubernetes.

Given the following database configuration we can derive the connections as follows:

```yaml
apiVersion: "acid.zalan.do/v1"
kind: postgresql
metadata:
  name: tadoku-contest-api-db
  labels:
    app: tadoku-contest-api-db
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
tadoku-contest-api-db

# Username, default superuser
postgres

# Password, secret name has the following pattern: `$username.$database_name.credentials.postgresql.acid.zalan.do`
$ kubectl get secret postgres.tadoku-contest-api-db.credentials.postgresql.acid.zalan.do -o json | jq -r .data.password | base64 --decode`

# Database name, derived from `spec.databases.$database_name`
tadoku
```

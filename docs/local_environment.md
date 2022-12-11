# Local Development Environment

Tadoku is made up of several services working together. It can be quite difficult to set up a local development environment with all the required services linked up together. This is a requirement for anyone to be productive in this project, and is also why we've provided a development environment for you.

We use [Tilt](https://tilt.dev/) to spin up a local Kubernetes cluster with all our backend services & dependencies. We've decided to leave out the frontend packages from this environment for now, as it turned out to be quite resource intensive when doing so. Each frontend will have some sort of development mode included which is configured to connect to this environment.

## Getting Started

1. Install [Helm](https://helm.sh/docs/intro/install/).
2. Install [Bazel](https://docs.bazel.build/bazel-overview.html).
3. Install [Tilt](https://docs.tilt.dev/install.html).
4. Read the [Getting Started Tutorial](https://docs.tilt.dev/tutorial.html) for Tilt to get familiar with it.
5. Run `$ tilt up` in the root of this repository.
6. Some services will have a database seed script, these can be manually triggered from within Tilt when needed.
7. Access the cluster from `http://langlog.be`, a domain reserved to serve a local dev instance of Tadoku.

## Can't connect connect to service/database

It's possible that the containers for a particular service or database have been shut down due resource constraints. In this case you can restart the service manually from the Tilt dashboard. If a database is unreachable it might be useful to restart the Tilt cluster.

## Connecting to a database within Tilt

We have configured a [pgweb instance](https://github.com/sosedoff/pgweb) for access to our PostgreSQL databases. The connection details can be found from Kubernetes.
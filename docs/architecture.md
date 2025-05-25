# Tadoku Architecture

## Overview

The Tadoku app consists of the following backend services & frontends:

### Backend services

- [immersion-api](/services/immersion-api.md)
- [content-api](/services/content-api.md)
- [Ory Kratos](https://github.com/ory/kratos)

### Frontends

- [webv2](/frontend/webv2.md)
- [auth](/frontend/auth.md)

### Infrastructure

- [Kong gateway](https://docs.konghq.com/gateway/latest/): ingress for all Traffic into the Kubernetes cluster
- [Ory Oathkeeper](https://github.com/ory/oathkeeper): identity & access proxy responsible for authorizing http traffic to the APIs.

## System Diagram

<img src="assets/architects.excalidraw.svg" alt="System diagram" style="max-width: 1000px;" />

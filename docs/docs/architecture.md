---
sidebar_position: 1
title: System Architecture
---

# Tadoku Architecture

## Overview

The Tadoku app consists of backend services and frontends deployed to Kubernetes. Developers use [DevCLI](https://github.com/tadoku/tadoku/blob/main/.dev/README.md) with the development cluster.

### Backend services

- [Tadoku API](https://github.com/tadoku/tadoku/tree/main/services/tadoku-api), including the public Immersion, [Content API](./services/content-api.md), Profile and Authorization namespaces
- [Ory Kratos](https://github.com/ory/kratos)

### Frontends

- [webv2](./frontend/webv2.md)
- [auth](./frontend/auth.md)

### Infrastructure

- [Kong gateway](https://docs.konghq.com/gateway/latest/): ingress for all Traffic into the Kubernetes cluster
- [Ory Oathkeeper](https://github.com/ory/oathkeeper): identity & access proxy responsible for authorizing http traffic to the APIs.

## Historical system diagram

This diagram predates the consolidation of Content and Profile API routes into Tadoku API.

![Historical system diagram](./assets/architects.excalidraw.svg)

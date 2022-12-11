# -*- mode: Python -*-

# Extensions
load('ext://helm_remote', 'helm_remote')

# Infra
# helm_remote('nats',
#             repo_name='nats',
#             repo_url='https://nats-io.github.io/k8s/helm/charts/',
#             set=['nats.image=synadia/nats-server:nightly', 'nats.jetstream.enabled=true'])
# helm_remote('nack',
#             repo_name='nats',
#             repo_url='https://nats-io.github.io/k8s/helm/charts/',
#             set=['jetstream.nats.url=nats://nats:4222'])

include('./infra/dev/gateway/Tiltfile')
include('./infra/dev/ory/Tiltfile')
include('./infra/dev/postgres/Tiltfile')
k8s_yaml('./infra/dev/tools/pgweb.yaml')

# Services

include('./services/reading-contest-api/Tiltfile')
include('./services/blog-api/Tiltfile')
include('./services/identity-api/Tiltfile')

# Frontend
# include('./frontend/Tiltfile')
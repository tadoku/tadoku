# -*- mode: Python -*-

# Extensions
load('ext://helm_remote', 'helm_remote')

# Infra
helm_remote('postgres-operator',
            repo_name='commonground',
            repo_url='https://charts.commonground.nl/')
helm_remote('nats',
            repo_name='nats',
            repo_url='https://nats-io.github.io/k8s/helm/charts/',
            set=['nats.image=synadia/nats-server:nightly', 'nats.jetstream.enabled=true'])
helm_remote('nack',
            repo_name='nats',
            repo_url='https://nats-io.github.io/k8s/helm/charts/',
            set=['jetstream.nats.url=nats://nats:4222'])

include('./gateway/Tiltfile')

# Tools
k8s_yaml('./tools/deployments/pgweb.yaml')
k8s_resource('pgweb', port_forwards=9000)

# Services

include('./services/reading-contest-api/Tiltfile')
include('./services/blog-api/Tiltfile')
include('./services/identity-api/Tiltfile')

# Frontend

# include('./frontend/web/Tiltfile')

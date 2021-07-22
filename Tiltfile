# -*- mode: Python -*-

# Extensions
load('ext://helm_remote', 'helm_remote')

# Infra
helm_remote('postgres-operator',
            repo_name='commonground',
            repo_url='https://charts.commonground.nl/')

# Tools
k8s_yaml('./tools/deployments/pgweb.yaml')
k8s_resource('pgweb', port_forwards=9000)

# Services

include('./services/tadoku-contest-api/Tiltfile')
include('./services/blog-api/Tiltfile')

# Frontends

# include('./frontends/web/Tiltfile')

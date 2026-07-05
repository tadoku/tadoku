# -*- mode: Python -*-

local_k8s_context = os.getenv('TADOKU_LOCAL_K8S_CONTEXT', '') or 'orbstack'
shared_k8s_context = 'dev-lab'

allow_k8s_contexts([local_k8s_context, shared_k8s_context])

if k8s_context() == shared_k8s_context:
    # Shared lab cluster: images are pushed to the homelab dev registry.
    default_registry('registry.dev.lab')
else:
    # Local cluster: images stay in the local docker daemon.
    # Phase 5: local-cluster bootstrap (postgres operator, etc.) goes here.
    pass

# Infra
include('./infra/dev/gateway/Tiltfile')
include('./infra/dev/ory/Tiltfile')
include('./infra/dev/postgres/Tiltfile')
include('./infra/dev/valkey/Tiltfile')
include('./infra/dev/tools/Tiltfile')

# Services

include('./services/immersion-api/Tiltfile')
include('./services/content-api/Tiltfile')
include('./services/profile-api/Tiltfile')
include('./services/authz-api/Tiltfile')

# Frontend
include('./frontend/Tiltfile')

# Private
# private_infra_path = '../tadoku-private/Tiltfile'
# if os.path.exists(private_infra_path):
#     include(private_infra_path)

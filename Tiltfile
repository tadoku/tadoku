# -*- mode: Python -*-

allow_k8s_contexts('orbstack')

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

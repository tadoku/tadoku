# -*- mode: Python -*-

load('./k8s/dev/config/Tiltfile', 'dev_registry', 'local_k8s_context', 'shared_k8s_context')

allow_k8s_contexts([local_k8s_context, shared_k8s_context])

if k8s_context() == shared_k8s_context:
    # Shared cluster: images are pushed to the registry configured in tilt_config.json.
    default_registry(dev_registry())
else:
    # Local cluster: images stay in the local docker daemon.
    pass

include('./k8s/dev/Tiltfile')

# Private
private_infra_path = os.getenv('TADOKU_PRIVATE_INFRA_TILTFILE', '')
if private_infra_path and os.path.exists(private_infra_path):
    include(private_infra_path)

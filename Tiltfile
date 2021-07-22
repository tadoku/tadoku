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

# -----------------------------
# blog
# -----------------------------

# Server container
watch_file('./services/blog-api/deployments/api.yaml')
k8s_yaml(local('bazel run //services/blog-api/deployments:api'))

custom_build(
  ref='blog-api-image',
  command=(
    'bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 {image_target} -- --norun && ' +
    'docker tag {bazel_image} $EXPECTED_REF').format(
      image_target='//services/blog-api:image',
      bazel_image='bazel/services/blog-api:image'
    ),
  deps=['services/blog-api'],
)

k8s_resource('blog-api', port_forwards=8001)

# -----------------------------
# tadoku-web
# -----------------------------

# include('./frontends/web/Tiltfile')

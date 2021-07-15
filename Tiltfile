# -*- mode: Python -*-

# Extensions
load('ext://helm_remote', 'helm_remote')

# Infra
helm_remote('postgres-operator',
            repo_name='commonground',
            repo_url='https://charts.commonground.nl/')

# Services

# -----------------------------
# tadoku-contest-api
# -----------------------------

# Server container
watch_file('./services/tadoku-contest-api/deployments/api.yaml')
k8s_yaml(local('bazel run //services/tadoku-contest-api/deployments:api'))

custom_build(
  ref='tadoku-contest-api-image',
  command=(
    'bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 {image_target} -- --norun && ' +
    'docker tag {bazel_image} $EXPECTED_REF').format(
      image_target='//services/tadoku-contest-api/cmd/server:base_image',
      bazel_image='bazel/services/tadoku-contest-api/cmd/server:base_image'
    ),
  deps=['services/tadoku-contest-api'],
)

k8s_resource('tadoku-contest-api', port_forwards=8000)

# Migration job
watch_file('./services/tadoku-contest-api/deployments/migrate.yaml')
k8s_yaml(local('bazel run //services/tadoku-contest-api/deployments:migrate'))

custom_build(
  ref='tadoku-contest-api-migrate-image',
  command=(
    'bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 {image_target} -- --norun && ' +
    'docker tag {bazel_image} $EXPECTED_REF').format(
      image_target='//services/tadoku-contest-api/cmd/migrate:image',
      bazel_image='bazel/services/tadoku-contest-api/cmd/migrate:image'
    ),
  deps=['services/tadoku-contest-api/interfaces/repositories/migrations'],
)

k8s_resource('tadoku-contest-api-migration-job', resource_deps=['tadoku-contest-api'])

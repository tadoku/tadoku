# -*- mode: Python -*-

# Extensions
load('ext://helm_remote', 'helm_remote')

# Services

# -----------------------------
# tadoku-contest-api
# -----------------------------
helm_remote('postgresql',
            repo_name='bitnami',
            repo_url='https://charts.bitnami.com/bitnami',
            values='./services/tadoku-contest-api/deployments/postgres-values.yaml')
watch_file('./services/tadoku-contest-api/deployments/tadoku-contest-api.yaml')
k8s_yaml(local('bazel run //services/tadoku-contest-api/deployments:tadoku-contest-api'))
custom_build(
  ref='tadoku-contest-api-image',
  command=(
    'bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 {image_target} -- --norun && ' +
    'docker tag {bazel_image} $EXPECTED_REF').format(
      image_target='services/tadoku-contest-api/deployments/BUILD',
      bazel_image='services/tadoku-contest-api/deployments/BUILD'
    ),
  deps=['main.go', 'web'],
)
k8s_resource('tadoku-contest-api', port_forwards=8000)

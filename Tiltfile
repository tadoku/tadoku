# -*- mode: Python -*-

# Use Bazel to generate the Kubernetes YAML
watch_file('./services/tadoku-contest-api/deployments/tadoku-contest-api.yaml')
k8s_yaml(local('bazel run //services/tadoku-contest-api/deployments:tadoku-contest-api'))

# Use Bazel to build the image

# The go_image BUILD rule
image_target='//:tadoku-contest-api-image'

# Where go_image puts the image in Docker (bazel/path/to/target:name)
bazel_image='bazel:tadoku-contest-api-image'

custom_build(
  ref='tadoku-contest-api-image',
  command=(
    'bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 {image_target} -- --norun && ' +
    'docker tag {bazel_image} $EXPECTED_REF').format(image_target=image_target, bazel_image=bazel_image),
  deps=['main.go', 'web'],
)

k8s_resource('tadoku-contest-api', port_forwards=8000)

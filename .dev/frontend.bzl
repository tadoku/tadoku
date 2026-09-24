load("@rules_oci//oci:defs.bzl", "oci_image", "oci_push")
load("@rules_pkg//pkg:mappings.bzl", "pkg_attributes", "pkg_files", "strip_prefix")
load("@rules_pkg//pkg:tar.bzl", "pkg_tar")
load("//.dev:deployable.bzl", "dev_deployable")

# Image startup and dev-cli's initial dependency sync may overlap. BusyBox flock
# serializes writes to node_modules and releases the lock if either process exits.
_DEV_INSTALL = [
    "flock",
    "/tmp/tadoku-pnpm-install.lock",
    "pnpm",
    "install",
    "--frozen-lockfile",
    "--network-concurrency=4",
    "--child-concurrency=1",
]

def frontend_dev(name, host, readiness_path = "/"):
    """One app source graph, OCI runtime and dev-cli contract."""
    command = " ".join(_DEV_INSTALL) + " && exec pnpm --filter " + name + " dev"

    native.filegroup(
        name = "%s_sources" % name,
        srcs = native.glob(
            [
                "apps/%s/**" % name,
                "packages/ui/**",
                "package.json",
                "pnpm-lock.yaml",
                "pnpm-workspace.yaml",
            ],
            exclude = [
                "**/node_modules/**",
                "**/.next/**",
                "**/.env*",
                "**/dist/**",
                "**/coverage/**",
                "**/test-results/**",
                "**/playwright-report/**",
                "**/*.tsbuildinfo",
                "**/next-env.d.ts",
            ],
        ),
    )

    pkg_files(
        name = "%s_dev_files" % name,
        srcs = [":%s_sources" % name],
        attributes = pkg_attributes(
            gid = 1000,
            mode = "0644",
            uid = 1000,
        ),
        prefix = "/app",
        strip_prefix = strip_prefix.from_pkg(),
    )

    pkg_tar(
        name = "%s_dev_layer" % name,
        srcs = [":%s_dev_files" % name],
    )

    oci_image(
        name = "%s_dev_image" % name,
        base = "@tadoku_dev_node_linux_amd64",
        cmd = [command],
        entrypoint = [
            "/bin/sh",
            "-ec",
        ],
        tars = [
            ":pnpm_dev_layer",
            ":%s_dev_layer" % name,
        ],
        user = "1000:1000",
        workdir = "/app",
    )

    oci_push(
        name = "%s_dev_push" % name,
        image = ":%s_dev_image" % name,
    )

    dev_deployable(
        name = "%s_dev" % name,
        metadata = json.encode({
            "name": "%s" % name,
            "kind": "frontend",
            "namespace": "tdk-dev-frontend-%s" % name,
            "buildTarget": "//frontend:%s_dev_image" % name,
            "imageTarget": "//frontend:%s_dev_image" % name,
            "imageName": "%s" % name,
            "pushTarget": "//frontend:%s_dev_push" % name,
            "port": 3000,
            "readinessPath": readiness_path,
            "publicPath": "/",
            "publicHost": host,
            "baseService": {
                "name": "frontend-%s" % name,
                "namespace": "tdk-dev-frontend-%s" % name,
                "port": 3000,
            },
            "devContainer": "%s" % name,
            "devCommand": [
                "/bin/sh",
                "-ec",
                command,
            ],
            "sourceRoots": [
                "frontend/apps/%s" % name,
                "frontend/packages/ui",
                "frontend/package.json",
                "frontend/pnpm-lock.yaml",
                "frontend/pnpm-workspace.yaml",
            ],
            "syncPaths": [
                "frontend/apps/%s" % name,
                "frontend/packages/ui",
                "frontend/package.json",
                "frontend/pnpm-lock.yaml",
                "frontend/pnpm-workspace.yaml",
            ],
            "syncRoot": "/app",
            "syncStripPrefix": "frontend",
            "syncExcludes": [
                "*.tsbuildinfo",
                "next-env.d.ts",
            ],
            "dependencyPaths": [
                "frontend/apps/%s/package.json" % name,
                "frontend/packages/ui/package.json",
                "frontend/package.json",
                "frontend/pnpm-lock.yaml",
                "frontend/pnpm-workspace.yaml",
            ],
            "dependencyCommand": _DEV_INSTALL,
        }),
        workload_template = "//.dev:%s.yaml" % name,
        deps = [
            ":%s_dev_image" % name,
            ":%s_sources" % name,
        ],
    )

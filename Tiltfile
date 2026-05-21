# -*- mode: Python -*-

# Tiltfile for Insights Service development
# Run with: task dev (which pre-builds binaries then runs tilt up)

# Configuration
allow_k8s_contexts(['kind-insights-dev', 'kind-kind'])

# Version is set via git describe or defaults to dev
version = str(local('git describe --tags --always --dirty 2>/dev/null || echo "0.0.0-dev"', quiet=True)).strip()
ldflags = '-X github.com/datum-cloud/insights/cmd/apiserver/app.Version=' + version

# Build controller docker image using custom_build to ensure binary exists
custom_build(
    'insights-controller',
    'CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin/manager ./cmd/controller && docker build --no-cache -f Dockerfile.dev -t $EXPECTED_REF .',
    deps=['cmd/controller', 'internal', 'pkg', 'go.mod', 'go.sum'],
)

# Build apiserver docker image using custom_build
custom_build(
    'insights-apiserver',
    'CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "' + ldflags + '" -o bin/insights-apiserver cmd/apiserver/main.go && docker build --no-cache -f Dockerfile.apiserver -t $EXPECTED_REF .',
    deps=['cmd/apiserver', 'pkg/apis', 'pkg/generated', 'internal/apiserver', 'internal/registry', 'go.mod', 'go.sum'],
)

# Deploy the controller
k8s_yaml(kustomize('config/controller'))

# Deploy the apiserver (includes etcd) - allow_duplicates for shared namespace
k8s_yaml(kustomize('config/apiserver'), allow_duplicates=True)

# Apply auth-reader RoleBinding separately (must be in kube-system, not affected by kustomize namespace)
k8s_yaml('config/apiserver/auth-reader-rolebinding.yaml')

# Configure the controller deployment
k8s_resource(
    'insights-controller',
    port_forwards=['8080:8080', '8081:8081'],
    labels=['controller'],
)

# Configure etcd for apiserver
k8s_resource(
    'etcd',
    labels=['apiserver'],
)

# Configure the apiserver deployment
k8s_resource(
    'insights-apiserver',
    port_forwards=['8443:8443'],
    resource_deps=['etcd'],
    labels=['apiserver'],
)

# Deploy sample policies
k8s_yaml(kustomize('config/samples'), allow_duplicates=True)

# Configure samples as a resource that depends on apiserver
k8s_resource(
    objects=['deployment-best-practices:insightpolicy'],
    new_name='samples',
    resource_deps=['insights-apiserver'],
    labels=['samples'],
)

# Build UI docker image
docker_build(
    'insights-ui',
    context='./ui',
    dockerfile='./ui/Dockerfile',
    live_update=[
        sync('./ui/src', '/app/src'),
        sync('./ui/public', '/app/public'),
    ]
)

# Deploy the UI
k8s_yaml(kustomize('ui/deploy'), allow_duplicates=True)

# Configure the UI deployment with port-forward
k8s_resource(
    'insights-ui',
    port_forwards=['3000:3000'],
    resource_deps=['insights-apiserver'],
    labels=['ui'],
)

# Samples - apply after apiserver is ready
# Use: kubectl apply -f config/samples/ after the apiserver is running

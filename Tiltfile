# -------------------------------
# 🏗 Cluster Bootstrap (Kind)
# -------------------------------
local_resource(
    'init_cluster',
    '''
    if ! kind get clusters | grep -q ratelimiter; then
      kind create cluster --name ratelimiter
    fi

    kubectl config use-context kind-ratelimiter

    # Wait for nodes to be ready
    for i in $(kubectl get nodes -o name); do
      kubectl wait --for=condition=Ready "$i" --timeout=60s || exit 1
    done
    ''',
    deps=[],
    allow_parallel=False
)

# -------------------------------
# 🐳 Docker build for operator
# -------------------------------
docker_build('nginx-ratelimits-operator', 'src', dockerfile='src/Dockerfile')

# -------------------------------
# 📦 Helm Deploy for Operator
# -------------------------------
k8s_yaml(local('helm template -f manifests/values.yaml manifests'))
# -------------------------------
# 📦 Load supporting test YAMLs
# -------------------------------
k8s_yaml([
    'test/namespace.yaml',
    'test/demo.yaml',
    'test/redis.yaml',
    'test/ratelimits.yaml',
    'test/svc.yaml',
])

# -------------------------------
# ⚙️ Define Kubernetes resource
# -------------------------------
k8s_resource(
    'ratelimits-operator',
    resource_deps=['init_cluster']
)

# -------------------------------
# 🌐 Port forward my-app
# -------------------------------

k8s_resource(
    'my-app',
    port_forwards='3000:80'
)


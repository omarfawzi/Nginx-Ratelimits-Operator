<div align="center">

# 🎛️ Nginx Ratelimits Operator

**A Kubernetes operator that injects an Nginx sidecar to enforce dynamic rate limiting rules.**

![Test Status](https://github.com/omarfawzi/Nginx-Ratelimits-Operator/actions/workflows/ci.yml/badge.svg)

---

Watches custom `RateLimits` resources and automatically mutates your deployments with the proper Nginx sidecar configuration.  
Works in tandem with the [Nginx Ratelimits Proxy](https://github.com/omarfawzi/Nginx-Ratelimits-Proxy) to apply request quotas to any Kubernetes workload.

</div>

---

## ✨ Features

| Feature | Description |
|---|---|
| **Custom Resource Definition** | Declarative `RateLimits` CRD for managing rate limit rules as native Kubernetes objects |
| **Automatic Sidecar Injection** | Mutates matching deployments to inject (or remove) the Nginx rate limiter proxy sidecar |
| **Helm Chart** | One-command installation via Helm |
| **Tilt Dev Environment** | Local Kind cluster with hot-reload for iterative development |
| **Example Manifests** | Ready-to-use samples in `test/` for quick testing |

---

## 🚀 Install via Helm

```bash
helm repo add nginx-ratelimits https://omarfawzi.github.io/Nginx-Ratelimits-Operator
helm install nginx-ratelimits-operator nginx-ratelimits/nginx-ratelimits-operator --version 1.5.0
```

---

## 📖 Usage

Apply a `RateLimits` resource and the operator will automatically inject the Nginx rate limiter proxy sidecar into matching pods:

```yaml
apiVersion: nginx.ratelimiter/v1alpha1
kind: RateLimits
metadata:
  name: my-app-ratelimits
  namespace: my-app
spec:
  selector:
    matchLabels:
      app: my-app

  env:
    UPSTREAM_HOST: "localhost"
    UPSTREAM_PORT: "5678"
    UPSTREAM_TYPE: "http"
    CACHE_PROVIDER: "redis"
    CACHE_HOST: "redis.default.svc.cluster.local"
    CACHE_PORT: "6379"
    REMOTE_IP_KEY: "remote_addr"

  rateLimits:
    ignoredSegments:
      users:
        - admin
      ips:
        - 127.0.0.1
      urls:
        - /v1/ping

    rules:
      /:
        ips:
          0.0.0.0/0:
            limit: 1
            window: 60
      /v1:
        users:
          user2:
            limit: 50
            window: 60
        ips:
          192.168.1.1:
            limit: 200
            window: 60
      ^/v2/[0-9]$:
        users:
          user3:
            flowRate: 10
            limit: 30
            window: 60
```

### How It Works

```mermaid
graph LR
    CRD(["📄 RateLimits CR"]):::crdStyle
    Operator["🎛️ Operator"]:::operatorStyle
    Deployment["📦 Deployment"]:::deployStyle
    Pod["🟢 Pod + Nginx Sidecar"]:::podStyle

    CRD -- "① Watch" --> Operator
    Operator -- "② Mutate" --> Deployment
    Deployment -- "③ Reconcile" --> Pod

    classDef crdStyle fill:#e3f2fd,stroke:#1565c0,stroke-width:2px,color:#0d47a1,font-weight:bold
    classDef operatorStyle fill:#fff3e0,stroke:#ef6c00,stroke-width:2px,color:#e65100,font-weight:bold
    classDef deployStyle fill:#f3e5f5,stroke:#7b1fa2,stroke-width:2px,color:#4a148c,font-weight:bold
    classDef podStyle fill:#e8f5e9,stroke:#2e7d32,stroke-width:2px,color:#1b5e20,font-weight:bold
```

| Field | Description |
|---|---|
| `spec.selector` | Label selector to match target deployments for sidecar injection |
| `spec.env` | Environment variables passed to the Nginx proxy sidecar (see [Proxy docs](https://github.com/omarfawzi/Nginx-Ratelimits-Proxy#environment-variables)) |
| `spec.rateLimits` | Rate limit rules and ignored segments (see [Rule format](https://github.com/omarfawzi/Nginx-Ratelimits-Proxy#rate-limit-rules)) |

> [!NOTE]
> Deleting the `RateLimits` resource will automatically remove the sidecar from affected deployments.

---

## 🛠 Development

[Tilt](https://docs.tilt.dev/install.html) creates a local Kind cluster and deploys the operator for iterative development with hot-reload:

```bash
# 1. Install Tilt — https://docs.tilt.dev/install.html
# 2. Start the dev environment
tilt up
```

Tilt will:
- Spin up a local Kind cluster
- Build the operator image
- Install the Helm chart
- Apply the sample manifests from `test/`

> See the `test/` directory for example workloads and rate limit definitions.

---

<div align="center">

**MIT License** · Powered by [Nginx Ratelimits Proxy](https://github.com/omarfawzi/Nginx-Ratelimits-Proxy)

</div>

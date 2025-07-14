![Test Status](https://github.com/omarfawzi/Nginx-Ratelimits-Operator/actions/workflows/ci.yml/badge.svg)

# Nginx Ratelimits Operator

A Kubernetes operator that injects an Nginx sidecar to enforce dynamic rate limiting rules. It watches custom `RateLimits` resources and automatically updates your application deployments with the proper Nginx configuration.

This operator works in tandem with the [Nginx Ratelimits Proxy](https://github.com/omarfawzi/Nginx-Ratelimits-Proxy) project to apply request quotas to any Kubernetes workload.

## Features

- Custom resource definition for managing rate limits
- Automatic sidecar injection and removal
- Example manifests for quick testing
- Helm chart for easy installation
- Development environment powered by Tilt

## Install via Helm

```bash
helm install nginx-ratelimits-operator oci://ghcr.io/omarfawzi/nginx-ratelimits-operator --version 1.1.0
```
## Development

Tilt can create a local Kind cluster and deploy the operator for iterative development:

1. Install [Tilt](https://docs.tilt.dev/install.html).
2. Run `tilt up` from the repository root.
3. Tilt builds the operator image, installs the Helm chart and applies the sample manifests in `test/`.

```bash
tilt up
```

See the `test/` directory for example workloads and rate limit definitions.

## Tests
Apply the sample `RateLimits` from the `test` directory to see the operator in action.

package controllers

// Ensure these constants are declared somewhere shared
const (
	sidecarName        = "rl-proxy"
	sidecarConfigMap   = "rl-config"
	sidecarImage       = "ghcr.io/omarfawzi/nginx-ratelimiter-proxy:kube-master"
	sidecarHash        = "rl-operator/hash"
	sidecarMountPath   = "/usr/local/openresty/nginx/lua/ratelimits.yaml"
	sidecarSubPath     = "ratelimits.yaml"
	sidecarPort        = 80
	selectorAnnotation = "rl-operator/last-selector"
)

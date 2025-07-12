package internal

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	v1 "ratelimits-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *RateLimitsReconciler) needsSidecarUpdate(deploy *appsv1.Deployment, rl *v1.RateLimits) bool {
	data := map[string]interface{}{
		"ratelimits": rl.Spec.RateLimits,
		"env":        rl.Spec.Env,
	}
	hashBytes, _ := json.Marshal(data)
	hash := fmt.Sprintf("%x", sha256.Sum256(hashBytes))

	if deploy.Spec.Template.Annotations == nil {
		deploy.Spec.Template.Annotations = map[string]string{}
	}

	return !hasSidecarInTemplate(deploy) || deploy.Spec.Template.Annotations[sidecarHash] != hash
}

func (r *RateLimitsReconciler) updateDeploymentHash(deploy *appsv1.Deployment, rl *v1.RateLimits) {
	data := map[string]interface{}{
		"ratelimits": rl.Spec.RateLimits,
		"env":        rl.Spec.Env,
	}
	hashBytes, _ := json.Marshal(data)
	hash := fmt.Sprintf("%x", sha256.Sum256(hashBytes))

	if deploy.Spec.Template.Annotations == nil {
		deploy.Spec.Template.Annotations = map[string]string{}
	}

	deploy.Spec.Template.Annotations[sidecarHash] = hash
}

func injectSideCar(logger logr.Logger, deploy *appsv1.Deployment, rl v1.RateLimits) {
	requiredEnv := []string{
		"UPSTREAM_PORT", "UPSTREAM_HOST", "UPSTREAM_TYPE",
		"CACHE_HOST", "CACHE_PORT", "CACHE_PROVIDER",
		"CACHE_PREFIX", "REMOTE_IP_KEY",
	}

	envVars := map[string]string{
		"CACHE_PREFIX": rl.Namespace,
	}
	for k, v := range rl.Spec.Env {
		envVars[k] = v
	}

	var missing []string
	for _, key := range requiredEnv {
		if _, ok := envVars[key]; !ok {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		logger.Error(nil, "Missing required environment variables for sidecar", "missing", missing)
		return
	}

	var env []corev1.EnvVar
	for k, v := range envVars {
		env = append(env, corev1.EnvVar{Name: k, Value: v})
	}

	sidecar := corev1.Container{
		Name:  sidecarName,
		Image: sidecarImage,
		Env:   env,
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("100m"),
				corev1.ResourceMemory: resource.MustParse("200Mi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("200m"),
				corev1.ResourceMemory: resource.MustParse("200Mi"),
			},
		},
		Ports: []corev1.ContainerPort{{
			ContainerPort: sidecarPort,
			Name:          sidecarName,
			Protocol:      corev1.ProtocolTCP,
		}},
		VolumeMounts: []corev1.VolumeMount{
			{Name: sidecarConfigMap, MountPath: sidecarMountPath, SubPath: sidecarSubPath},
		},
	}

	if idx := findContainerIndex(deploy, sidecarName); idx >= 0 {
		deploy.Spec.Template.Spec.Containers[idx] = sidecar
	} else {
		deploy.Spec.Template.Spec.Containers = append(deploy.Spec.Template.Spec.Containers, sidecar)
	}

	addConfigVolumes(&deploy.Spec.Template)
}

func (r *RateLimitsReconciler) removeSidecarFromOldMatches(ctx context.Context, ns string, oldSel, newSel labels.Selector) {
	logger := log.FromContext(ctx)

	var pods corev1.PodList
	if err := r.List(ctx, &pods, &client.ListOptions{Namespace: ns, LabelSelector: oldSel}); err != nil {
		return
	}

	for _, pod := range pods.Items {
		if newSel.Matches(labels.Set(pod.Labels)) {
			continue
		}

		rsOwner := metav1.GetControllerOf(&pod)
		if rsOwner == nil || rsOwner.Kind != "ReplicaSet" {
			continue
		}

		var rs appsv1.ReplicaSet
		if err := r.Get(ctx, types.NamespacedName{Name: rsOwner.Name, Namespace: pod.Namespace}, &rs); err != nil {
			continue
		}

		deployOwner := metav1.GetControllerOf(&rs)
		if deployOwner == nil || deployOwner.Kind != "Deployment" {
			continue
		}

		var deploy appsv1.Deployment
		if err := r.Get(ctx, types.NamespacedName{Name: deployOwner.Name, Namespace: pod.Namespace}, &deploy); err != nil {
			continue
		}

		if hasSidecarInTemplate(&deploy) {
			removeSidecarContainer(&deploy)
			delete(deploy.Spec.Template.Annotations, sidecarHash)
			if err := r.Update(ctx, &deploy); err != nil {
				logger.Error(err, "Failed to update Deployment to remove sidecar", "deployment", deploy.Name)
			}
		}
	}
}

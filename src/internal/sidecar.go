package internal

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	v1 "ratelimits-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func (r *RateLimitsReconciler) needsStatefulSetSidecarUpdate(sts *appsv1.StatefulSet, rl *v1.RateLimits) bool {
	data := map[string]interface{}{
		"ratelimits": rl.Spec.RateLimits,
		"env":        rl.Spec.Env,
	}
	hashBytes, _ := json.Marshal(data)
	hash := fmt.Sprintf("%x", sha256.Sum256(hashBytes))

	if sts.Spec.Template.Annotations == nil {
		sts.Spec.Template.Annotations = map[string]string{}
	}

	return !hasSidecarInStatefulSet(sts) || sts.Spec.Template.Annotations[sidecarHash] != hash
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

func (r *RateLimitsReconciler) updateStatefulSetHash(sts *appsv1.StatefulSet, rl *v1.RateLimits) {
	data := map[string]interface{}{
		"ratelimits": rl.Spec.RateLimits,
		"env":        rl.Spec.Env,
	}
	hashBytes, _ := json.Marshal(data)
	hash := fmt.Sprintf("%x", sha256.Sum256(hashBytes))

	if sts.Spec.Template.Annotations == nil {
		sts.Spec.Template.Annotations = map[string]string{}
	}

	sts.Spec.Template.Annotations[sidecarHash] = hash
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

func injectSideCarStatefulSet(logger logr.Logger, sts *appsv1.StatefulSet, rl v1.RateLimits) {
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

	if idx := findContainerIndexStatefulSet(sts, sidecarName); idx >= 0 {
		sts.Spec.Template.Spec.Containers[idx] = sidecar
	} else {
		sts.Spec.Template.Spec.Containers = append(sts.Spec.Template.Spec.Containers, sidecar)
	}

	addConfigVolumes(&sts.Spec.Template)
}

func (r *RateLimitsReconciler) removeSidecarFromOldMatches(ctx context.Context, ns string, oldSel, newSel labels.Selector) {
	var pods corev1.PodList
	if err := r.List(ctx, &pods, &client.ListOptions{Namespace: ns, LabelSelector: oldSel}); err != nil {
		return
	}

	for _, pod := range pods.Items {
		if newSel.Matches(labels.Set(pod.Labels)) {
			continue
		}

		owner := metav1.GetControllerOf(&pod)
		if owner == nil {
			continue
		}

		switch owner.Kind {
		case "ReplicaSet":
			var rs appsv1.ReplicaSet
			if err := r.Get(ctx, types.NamespacedName{Name: owner.Name, Namespace: pod.Namespace}, &rs); err != nil {
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

			r.removeSidecarIfExists(ctx, deploy)
		case "StatefulSet":
			var sts appsv1.StatefulSet
			if err := r.Get(ctx, types.NamespacedName{Name: owner.Name, Namespace: pod.Namespace}, &sts); err != nil {
				continue
			}

			r.removeSidecarStatefulSetIfExists(ctx, sts)
		}
	}
}

func (r *RateLimitsReconciler) removeSidecarIfExists(ctx context.Context, deploy appsv1.Deployment) {
	logger := log.FromContext(ctx)

	if hasSidecarInTemplate(&deploy) {
		orig := deploy.DeepCopy()
		removeSidecarContainer(&deploy)
		delete(deploy.Spec.Template.Annotations, sidecarHash)
		if err := r.Patch(ctx, &deploy, client.MergeFrom(orig)); err != nil {
			if errors.IsConflict(err) {
				logger.Info("Skipping update due to conflict", "deployment", deploy.Name)
			} else {
				logger.Error(err, "Failed to update Deployment with sidecar", "deployment", deploy.Name)
			}
		}
	}
}

func (r *RateLimitsReconciler) removeSidecarStatefulSetIfExists(ctx context.Context, sts appsv1.StatefulSet) {
	logger := log.FromContext(ctx)

	if hasSidecarInStatefulSet(&sts) {
		orig := sts.DeepCopy()
		removeSidecarContainerStatefulSet(&sts)
		delete(sts.Spec.Template.Annotations, sidecarHash)
		if err := r.Patch(ctx, &sts, client.MergeFrom(orig)); err != nil {
			if errors.IsConflict(err) {
				logger.Info("Skipping update due to conflict", "statefulset", sts.Name)
			} else {
				logger.Error(err, "Failed to update StatefulSet with sidecar", "statefulset", sts.Name)
			}
		}
	}
}

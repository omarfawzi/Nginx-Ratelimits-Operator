package internal

import (
	"context"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"ratelimits-operator/api/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *RateLimitsReconciler) listSelectedPods(ctx context.Context, ns string, selector labels.Selector) (*corev1.PodList, error) {
	var podList corev1.PodList
	err := r.List(ctx, &podList, &client.ListOptions{Namespace: ns, LabelSelector: selector})
	return &podList, err
}

func (r *RateLimitsReconciler) reconcilePod(ctx context.Context, logger logr.Logger, pod *corev1.Pod, rl *v1alpha1.RateLimits) {
	if pod.DeletionTimestamp != nil {
		return
	}

	rsOwner := v1.GetControllerOf(pod)
	if rsOwner == nil || rsOwner.Kind != "ReplicaSet" {
		return
	}

	var rs appsv1.ReplicaSet
	if err := r.Get(ctx, types.NamespacedName{Name: rsOwner.Name, Namespace: pod.Namespace}, &rs); err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "Error getting ReplicaSet", "pod", pod.Name)
		}
		return
	}

	deployOwner := v1.GetControllerOf(&rs)
	if deployOwner == nil || deployOwner.Kind != "Deployment" {
		return
	}

	var deploy appsv1.Deployment
	if err := r.Get(ctx, types.NamespacedName{Name: deployOwner.Name, Namespace: pod.Namespace}, &deploy); err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "Error getting Deployment", "pod", pod.Name)
		}
		return
	}

	if r.needsSidecarUpdate(&deploy, rl) {
		orig := deploy.DeepCopy()
		injectSideCar(logger, &deploy, *rl)
		r.updateDeploymentHash(&deploy, rl)
		if err := r.Patch(ctx, &deploy, client.MergeFrom(orig)); err != nil {
			if errors.IsConflict(err) {
				logger.Info("Skipping update due to conflict", "deployment", deploy.Name)
			} else {
				logger.Error(err, "Failed to update Deployment with sidecar", "deployment", deploy.Name)
			}
		}
	}
}

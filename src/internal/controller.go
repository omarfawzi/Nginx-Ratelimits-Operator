package controllers

import (
	"context"
	"encoding/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	v1 "ratelimits-operator/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type RateLimitsReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *RateLimitsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var rateLimits v1.RateLimits
	if err := r.Get(ctx, req.NamespacedName, &rateLimits); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	selector, err := metav1.LabelSelectorAsSelector(&rateLimits.Spec.Selector)
	if err != nil {
		logger.Error(err, "Invalid selector in RateLimits")
		return ctrl.Result{}, nil
	}

	oldSelectorStr := rateLimits.Annotations[selectorAnnotation]
	currentSelectorStr := selectorToString(rateLimits.Spec.Selector)

	if oldSelectorStr != "" && oldSelectorStr != currentSelectorStr {
		var oldSelector metav1.LabelSelector
		if err := json.Unmarshal([]byte(oldSelectorStr), &oldSelector); err == nil {
			if oldSel, err := metav1.LabelSelectorAsSelector(&oldSelector); err == nil {
				r.removeSidecarFromOldMatches(ctx, rateLimits.Namespace, oldSel, selector)
			}
		}
	}

	pods, err := r.listSelectedPods(ctx, rateLimits.Namespace, selector)
	if err != nil {
		return ctrl.Result{}, err
	}

	for i := range pods.Items {
		r.reconcilePod(ctx, logger, &pods.Items[i], &rateLimits)
	}

	if err := r.ensureRateLimitConfigMap(ctx, &rateLimits); err != nil {
		return ctrl.Result{}, err
	}

	if rateLimits.Annotations == nil {
		rateLimits.Annotations = map[string]string{}
	}
	rateLimits.Annotations[selectorAnnotation] = currentSelectorStr
	if err := r.Update(ctx, &rateLimits); err != nil {
		logger.Error(err, "Failed to update selector annotation")
	}

	return ctrl.Result{}, nil
}

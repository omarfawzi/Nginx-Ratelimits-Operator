package internal

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	v1 "ratelimits-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *RateLimitsReconciler) mapDeploymentToCR(ctx context.Context, obj client.Object) []reconcile.Request {
	var crList v1.RateLimitsList
	if err := r.List(ctx, &crList, &client.ListOptions{Namespace: obj.GetNamespace()}); err != nil {
		return nil
	}

	var requests []reconcile.Request
	for _, cr := range crList.Items {
		selector, err := metav1.LabelSelectorAsSelector(&cr.Spec.Selector)
		if err != nil {
			continue
		}
		deploy, ok := obj.(*appsv1.Deployment)
		if !ok {
			continue
		}
		if selector.Matches(labels.Set(deploy.Spec.Template.Labels)) {
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace},
			})
		}
	}
	return requests
}

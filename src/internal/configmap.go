package internal

import (
	"context"
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "ratelimits-operator/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/yaml"
)

func (r *RateLimitsReconciler) ensureRateLimitConfigMap(ctx context.Context, rl *v1.RateLimits) error {
	cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: sidecarConfigMap, Namespace: rl.Namespace}}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, cm, func() error {
		raw, err := json.Marshal(rl.Spec.RateLimits)
		if err != nil {
			return err
		}
		yamlData, err := yaml.JSONToYAML(raw)
		if err != nil {
			return err
		}
		cm.Data = map[string]string{"ratelimits.yaml": string(yamlData)}
		return ctrl.SetControllerReference(rl, cm, r.Scheme)
	})

	return err
}

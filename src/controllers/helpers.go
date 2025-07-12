package controllers

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func findContainerIndex(deploy *appsv1.Deployment, name string) int {
	for i, c := range deploy.Spec.Template.Spec.Containers {
		if c.Name == name {
			return i
		}
	}
	return -1
}

func hasSidecarInTemplate(deploy *appsv1.Deployment) bool {
	return findContainerIndex(deploy, sidecarName) >= 0
}

func removeSidecarContainer(deploy *appsv1.Deployment) {
	var updated []corev1.Container
	for _, c := range deploy.Spec.Template.Spec.Containers {
		if c.Name != sidecarName {
			updated = append(updated, c)
		}
	}
	deploy.Spec.Template.Spec.Containers = updated
}

func addConfigVolumes(tmpl *corev1.PodTemplateSpec) {
	for _, v := range tmpl.Spec.Volumes {
		if v.Name == sidecarConfigMap {
			return
		}
	}
	tmpl.Spec.Volumes = append(tmpl.Spec.Volumes, corev1.Volume{
		Name: sidecarConfigMap,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: sidecarConfigMap},
			},
		},
	})
}

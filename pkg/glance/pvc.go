package glance

import (
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func PersistentVolumeClaims(instance *openstackv1beta1.Glance) []*corev1.PersistentVolumeClaim {
	var claims []*corev1.PersistentVolumeClaim

	for _, backend := range instance.Spec.Backends {
		pvcSpec := backend.PVC
		if pvcSpec == nil {
			continue
		}

		name := template.Combine(instance.Name, backend.Name)
		labels := template.AppLabels(instance.Name, AppLabel)

		pvc := template.PersistentVolumeClaim(name, labels, pvcSpec)
		pvc.Namespace = instance.Namespace

		claims = append(claims, &pvc)
	}

	return claims
}

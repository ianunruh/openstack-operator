package rally

import (
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func PersistentVolumeClaim(instance *openstackv1beta1.Rally) *corev1.PersistentVolumeClaim {
	name := template.Combine(instance.Name, "data")
	labels := template.AppLabels(instance.Name, AppLabel)

	pvc := template.PersistentVolumeClaim(name, labels, instance.Spec.Data)
	pvc.Namespace = instance.Namespace

	return &pvc
}

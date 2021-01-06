package controlplane

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func Glance(instance *openstackv1beta1.ControlPlane) *openstackv1beta1.Glance {
	spec := instance.Spec.Glance

	return &openstackv1beta1.Glance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "glance",
			Namespace: instance.Namespace,
		},
		Spec: spec,
	}
}

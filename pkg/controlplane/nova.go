package controlplane

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func Nova(instance *openstackv1beta1.ControlPlane) *openstackv1beta1.Nova {
	// TODO labels
	spec := instance.Spec.Nova

	return &openstackv1beta1.Nova{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nova",
			Namespace: instance.Namespace,
		},
		Spec: spec,
	}
}

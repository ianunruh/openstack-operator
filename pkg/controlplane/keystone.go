package controlplane

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func Keystone(instance *openstackv1beta1.ControlPlane) *openstackv1beta1.Keystone {
	// TODO labels
	spec := instance.Spec.Keystone

	spec.API.Ingress = ingressDefaults(spec.API.Ingress, instance, "keystone")

	return &openstackv1beta1.Keystone{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "keystone",
			Namespace: instance.Namespace,
		},
		Spec: spec,
	}
}

func DemoKeystoneUser(instance *openstackv1beta1.ControlPlane) *openstackv1beta1.KeystoneUser {
	return &openstackv1beta1.KeystoneUser{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo",
			Namespace: instance.Namespace,
		},
		Spec: openstackv1beta1.KeystoneUserSpec{
			Secret:  "demo-keystone",
			Project: "demo",
			Roles:   []string{"heat_stack_owner"},
		},
	}
}

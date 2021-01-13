package controlplane

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func Nova(instance *openstackv1beta1.ControlPlane) *openstackv1beta1.Nova {
	// TODO labels
	spec := instance.Spec.Nova

	spec.API.Ingress = ingressDefaults(spec.API.Ingress, instance, "nova")
	spec.Cells = novaCellDefaults(spec.Cells, instance)

	return &openstackv1beta1.Nova{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nova",
			Namespace: instance.Namespace,
		},
		Spec: spec,
	}
}

func novaCellDefaults(cells []openstackv1beta1.NovaCellSpec, instance *openstackv1beta1.ControlPlane) []openstackv1beta1.NovaCellSpec {
	out := make([]openstackv1beta1.NovaCellSpec, 0, len(cells))

	for _, cell := range cells {
		// TODO handle naming for multiple cells
		cell.NoVNCProxy.Ingress = ingressDefaults(cell.NoVNCProxy.Ingress, instance, "novnc")

		out = append(out, cell)
	}

	return out
}

package controlplane

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func Nova(instance *openstackv1beta1.ControlPlane) *openstackv1beta1.Nova {
	// TODO labels
	spec := instance.Spec.Nova

	spec.API.Ingress = ingressDefaults(spec.API.Ingress, instance, "nova")
	spec.API.NodeSelector = controllerNodeSelector(spec.API.NodeSelector, instance)

	spec.Cells = novaCellDefaults(spec.Cells, instance)

	spec.Conductor.NodeSelector = controllerNodeSelector(spec.Conductor.NodeSelector, instance)

	spec.Scheduler.NodeSelector = controllerNodeSelector(spec.Scheduler.NodeSelector, instance)

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

	for _, spec := range cells {
		spec.Conductor.NodeSelector = controllerNodeSelector(spec.Conductor.NodeSelector, instance)

		for name, compute := range spec.Compute {
			spec.Compute[name] = novaComputeDefaults(compute, spec, instance)
		}

		spec.Metadata.NodeSelector = controllerNodeSelector(spec.Metadata.NodeSelector, instance)

		// TODO handle naming for multiple cells
		spec.NoVNCProxy.Ingress = ingressDefaults(spec.NoVNCProxy.Ingress, instance, "novnc")
		spec.NoVNCProxy.NodeSelector = controllerNodeSelector(spec.NoVNCProxy.NodeSelector, instance)

		out = append(out, spec)
	}

	return out
}

func novaComputeDefaults(spec openstackv1beta1.NovaComputeSpec, cell openstackv1beta1.NovaCellSpec, instance *openstackv1beta1.ControlPlane) openstackv1beta1.NovaComputeSpec {
	spec.NodeSelector = computeNodeSelector(spec.NodeSelector, instance)

	return spec
}

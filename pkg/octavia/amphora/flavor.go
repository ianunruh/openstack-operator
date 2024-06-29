package amphora

import (
	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newFlavor(instance *openstackv1beta1.Octavia) *openstackv1beta1.NovaFlavor {
	// TODO make flavor opts configurable
	flavorDisk := 10

	return &openstackv1beta1.NovaFlavor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "c1-amphora",
			Namespace: instance.Namespace,
		},
		Spec: openstackv1beta1.NovaFlavorSpec{
			VCPUs: 2,
			RAM:   2048,
			Disk:  &flavorDisk,
		},
	}
}

package amphora

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	novaflavor "github.com/ianunruh/openstack-operator/pkg/nova/flavor"
)

func (b *bootstrap) EnsureFlavor(ctx context.Context) error {
	flavor := newFlavor(b.instance)
	controllerutil.SetControllerReference(b.instance, flavor, b.client.Scheme())
	if err := novaflavor.Ensure(ctx, b.client, flavor, b.log); err != nil {
		return err
	}
	novaflavor.AddReadyCheck(b.deps, flavor)

	if b.instance.Status.Amphora.FlavorID == flavor.Status.FlavorID {
		return nil
	}

	b.instance.Status.Amphora.FlavorID = flavor.Status.FlavorID
	if err := b.client.Status().Update(ctx, b.instance); err != nil {
		return err
	}

	return nil
}

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

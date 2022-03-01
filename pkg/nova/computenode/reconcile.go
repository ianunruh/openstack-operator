package computenode

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/hypervisors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func Reconcile(ctx context.Context, c client.Client, instance *openstackv1beta1.NovaComputeNode, compute *gophercloud.ServiceClient, log logr.Logger) error {
	hv, err := findHypervisor(instance.Spec.Node, compute)
	if err != nil {
		return err
	}

	syncStatus(instance, hv)
	if err := c.Status().Update(ctx, instance); err != nil {
		return err
	}

	return nil
}

func findHypervisor(hostname string, compute *gophercloud.ServiceClient) (*hypervisors.Hypervisor, error) {
	pages, err := hypervisors.List(compute, hypervisors.ListOpts{
		HypervisorHostnamePattern: &hostname,
	}).AllPages()
	if err != nil {
		return nil, err
	}

	current, err := hypervisors.ExtractHypervisors(pages)
	if err != nil {
		return nil, err
	}

	for _, hv := range current {
		if hv.HypervisorHostname == hostname {
			return &hv, nil
		}
	}

	return nil, errors.New("hypervisor not found")
}

func syncStatus(instance *openstackv1beta1.NovaComputeNode, hv *hypervisors.Hypervisor) {
	instance.Status.Hypervisor = &openstackv1beta1.NovaHypervisorStatus{
		Enabled:            hv.Status == "enabled",
		Up:                 hv.State == "up",
		HostIP:             hv.HostIP,
		HypervisorType:     hv.HypervisorType,
		RunningServerCount: hv.RunningVMs,
		TaskCount:          hv.CurrentWorkload,
		ServiceID:          hv.Service.ID,
	}
}

package neutron

import (
	"context"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	corev1 "k8s.io/api/core/v1"

	"github.com/ianunruh/openstack-operator/pkg/keystone"
)

func NewNetworkServiceClient(ctx context.Context, svcUser *corev1.Secret) (*gophercloud.ServiceClient, error) {
	client, err := keystone.CloudClient(svcUser)
	if err != nil {
		return nil, err
	}

	// pass through context from controller
	client.Context = ctx

	endpointOpts := gophercloud.EndpointOpts{
		Region:       string(svcUser.Data["OS_REGION_NAME"]),
		Availability: gophercloud.AvailabilityPublic,
	}

	return openstack.NewNetworkV2(client, endpointOpts)
}

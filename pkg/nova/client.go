package nova

import (
	"context"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	corev1 "k8s.io/api/core/v1"

	"github.com/ianunruh/openstack-operator/pkg/keystone"
)

func NewComputeServiceClient(ctx context.Context, svcUser *corev1.Secret) (*gophercloud.ServiceClient, error) {
	client, err := keystone.CloudClient(ctx, svcUser)
	if err != nil {
		return nil, err
	}

	return openstack.NewComputeV2(client, keystone.CloudEndpointOpts(svcUser))
}

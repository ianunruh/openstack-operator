package amphora

import (
	"context"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
)

func (b *bootstrap) EnsureNetwork(ctx context.Context) error {
	if len(b.instance.Status.Amphora.NetworkIDs) > 0 {
		// TODO check if current exists, otherwise recreate
		return nil
	}

	b.log.Info("Creating network", "name", amphoraNetworkName)
	network, err := networks.Create(b.network, networks.CreateOpts{
		Name: amphoraNetworkName,
	}).Extract()
	if err != nil {
		return err
	}

	b.log.Info("Creating subnet", "networkID", network.ID)
	_, err = subnets.Create(b.network, subnets.CreateOpts{
		NetworkID: network.ID,
		IPVersion: gophercloud.IPv4,
		CIDR:      b.instance.Spec.Amphora.ManagementCIDR,
	}).Extract()
	if err != nil {
		return err
	}

	b.instance.Status.Amphora.NetworkIDs = []string{network.ID}
	if err := b.client.Status().Update(ctx, b.instance); err != nil {
		return err
	}

	return nil
}

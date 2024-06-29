package amphora

import (
	"context"

	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/groups"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/rules"
)

func (b *bootstrap) EnsureSecurityGroup(ctx context.Context) error {
	if len(b.instance.Status.Amphora.SecurityGroupIDs) > 0 {
		return nil
	}

	b.log.Info("Creating security group", "name", "octavia-lb-mgmt")
	group, err := groups.Create(b.network, groups.CreateOpts{
		Name: "octavia-lb-mgmt",
	}).Extract()
	if err != nil {
		return err
	}

	// TODO reconcile each rule independently
	ruleOpts := []rules.CreateOpts{
		{Protocol: rules.ProtocolICMP},
		{Protocol: rules.ProtocolTCP, PortRangeMin: 22, PortRangeMax: 22},
		{Protocol: rules.ProtocolTCP, PortRangeMin: 9443, PortRangeMax: 9443},
	}
	for _, opts := range ruleOpts {
		opts.SecGroupID = group.ID
		opts.Direction = rules.DirIngress
		opts.EtherType = rules.EtherType4
		if err := rules.Create(b.network, opts).Err; err != nil {
			return err
		}
	}

	b.instance.Status.Amphora.SecurityGroupIDs = []string{group.ID}
	if err := b.client.Status().Update(ctx, b.instance); err != nil {
		return err
	}

	return nil
}

func (b *bootstrap) EnsureHealthSecurityGroup(ctx context.Context) error {
	if len(b.instance.Status.Amphora.HealthSecurityGroupIDs) > 0 {
		return nil
	}

	b.log.Info("Creating security group", "name", "octavia-lb-health-manager")
	group, err := groups.Create(b.network, groups.CreateOpts{
		Name: "octavia-lb-health-manager",
	}).Extract()
	if err != nil {
		return err
	}

	// TODO reconcile each rule independently
	ruleOpts := []rules.CreateOpts{
		{Protocol: rules.ProtocolUDP, PortRangeMin: 5555, PortRangeMax: 5555},
	}
	for _, opts := range ruleOpts {
		opts.SecGroupID = group.ID
		opts.Direction = rules.DirIngress
		opts.EtherType = rules.EtherType4
		if err := rules.Create(b.network, opts).Err; err != nil {
			return err
		}
	}

	b.instance.Status.Amphora.HealthSecurityGroupIDs = []string{group.ID}
	if err := b.client.Status().Update(ctx, b.instance); err != nil {
		return err
	}

	return nil
}

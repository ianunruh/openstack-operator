package network

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/external"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/provider"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func Reconcile(ctx context.Context, c client.Client, instance *openstackv1beta1.NeutronNetwork, client *gophercloud.ServiceClient, log logr.Logger) error {
	network, err := getNetwork(instance, client)
	if err != nil {
		if _, ok := err.(gophercloud.ErrDefault404); !ok {
			return fmt.Errorf("getting network: %w", err)
		}
	}

	if err := reconcileNetwork(ctx, c, instance, network, client, log); err != nil {
		return err
	}

	return nil
}

func getNetwork(instance *openstackv1beta1.NeutronNetwork, client *gophercloud.ServiceClient) (*networks.Network, error) {
	// fetch by ID
	if len(instance.Status.ProviderID) > 0 {
		network, err := networks.Get(client, instance.Status.ProviderID).Extract()
		if err != nil {
			if _, ok := err.(gophercloud.ErrDefault404); !ok {
				return nil, err
			}
		} else if network != nil {
			return network, nil
		}
	}

	// fetch by name
	name := networkName(instance)
	return findNetworkByName(name, client)
}

func findNetworkByName(name string, compute *gophercloud.ServiceClient) (*networks.Network, error) {
	// TODO filter by project if requested
	pages, err := networks.List(compute, networks.ListOpts{}).AllPages()
	if err != nil {
		return nil, err
	}

	current, err := networks.ExtractNetworks(pages)
	if err != nil {
		return nil, err
	}

	for _, network := range current {
		if network.Name == name {
			return &network, nil
		}
	}

	return nil, nil
}

func reconcileNetwork(ctx context.Context, c client.Client, instance *openstackv1beta1.NeutronNetwork, network *networks.Network, client *gophercloud.ServiceClient, log logr.Logger) error {
	name := networkName(instance)

	// TODO support looking up project by name
	projectID := instance.Spec.Project

	segments := make([]provider.Segment, 0, len(instance.Spec.Segments))
	for _, segment := range instance.Spec.Segments {
		segments = append(segments, provider.Segment{
			PhysicalNetwork: segment.PhysicalNetwork,
			NetworkType:     segment.NetworkType,
			SegmentationID:  segment.SegmentationID,
		})
	}

	var err error

	// create new network
	if network == nil {
		var opts networks.CreateOptsBuilder = networks.CreateOpts{
			Name:                  name,
			Description:           instance.Spec.Description,
			AdminStateUp:          instance.Spec.AdminStateUp,
			Shared:                instance.Spec.Shared,
			ProjectID:             projectID,
			AvailabilityZoneHints: instance.Spec.AvailabilityZoneHints,
		}

		opts = external.CreateOptsExt{
			External:          instance.Spec.External,
			CreateOptsBuilder: opts,
		}

		opts = provider.CreateOptsExt{
			Segments:          segments,
			CreateOptsBuilder: opts,
		}

		network, err = networks.Create(client, opts).Extract()
		if err != nil {
			return fmt.Errorf("creating network: %w", err)
		}
	}

	if networkChanged(instance, network) {
		var opts networks.UpdateOptsBuilder = networks.UpdateOpts{
			Name:         &name,
			Description:  &instance.Spec.Description,
			AdminStateUp: instance.Spec.AdminStateUp,
			Shared:       instance.Spec.Shared,
		}

		opts = external.UpdateOptsExt{
			External:          instance.Spec.External,
			UpdateOptsBuilder: opts,
		}

		network, err = networks.Update(client, network.ID, opts).Extract()
		if err != nil {
			return err
		}
	}

	// ensure provider ID present in status
	// TODO update provider condition based on current network status
	if instance.Status.ProviderID != network.ID {
		instance.Status.ProviderID = network.ID
		if err := c.Status().Update(ctx, instance); err != nil {
			return err
		}
	}

	return nil
}

func networkChanged(instance *openstackv1beta1.NeutronNetwork, network *networks.Network) bool {
	if network.Name != networkName(instance) {
		return true
	}

	if network.Description != instance.Spec.Description {
		return true
	}

	// TODO adminStateUp
	// TODO shared
	// TODO external

	return false
}

func networkName(instance *openstackv1beta1.NeutronNetwork) string {
	if instance.Spec.Name == "" {
		return instance.Name
	}
	return instance.Spec.Name
}

func Delete(instance *openstackv1beta1.NeutronNetwork, client *gophercloud.ServiceClient, log logr.Logger) error {
	network, err := getNetwork(instance, client)
	if err != nil {
		return err
	} else if network == nil {
		log.Info("Network not found for deletion", "name", instance.Name)
		return nil
	}

	log.Info("Deleting network", "name", instance.Name)
	if err := networks.Delete(client, network.ID).Err; err != nil {
		if _, ok := err.(gophercloud.ErrDefault404); !ok {
			return err
		}
		log.Info("Network not found on deletion", "name", instance.Name)
	}

	return nil
}

func Ensure(ctx context.Context, c client.Client, instance *openstackv1beta1.NeutronNetwork, log logr.Logger) error {
	intended := instance.DeepCopy()
	hash, err := template.ObjectHash(instance)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}

	if err := c.Get(ctx, client.ObjectKeyFromObject(instance), instance); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}

		template.SetAppliedHash(intended, hash)

		log.Info("Creating NeutronNetwork", "Name", instance.Name)
		return c.Create(ctx, instance)
	} else if !template.MatchesAppliedHash(instance, hash) {
		instance.Spec = intended.Spec
		template.SetAppliedHash(instance, hash)

		log.Info("Updating NeutronNetwork", "Name", instance.Name)
		return c.Update(ctx, instance)
	}

	return nil
}

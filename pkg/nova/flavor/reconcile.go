package flavor

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func Reconcile(ctx context.Context, c client.Client, instance *openstackv1beta1.NovaFlavor, compute *gophercloud.ServiceClient, log logr.Logger) error {
	flavor, err := getFlavor(instance, compute)
	if err != nil {
		if _, ok := err.(gophercloud.ErrDefault404); !ok {
			return fmt.Errorf("getting flavor: %w", err)
		}
	}

	if err := reconcileFlavor(ctx, c, instance, flavor, compute, log); err != nil {
		return err
	}

	return nil
}

func getFlavor(instance *openstackv1beta1.NovaFlavor, compute *gophercloud.ServiceClient) (*flavors.Flavor, error) {
	// fetch by ID
	if len(instance.Status.FlavorID) > 0 {
		flavor, err := flavors.Get(compute, instance.Status.FlavorID).Extract()
		if err != nil {
			if _, ok := err.(gophercloud.ErrDefault404); !ok {
				return nil, err
			}
		} else if flavor != nil {
			return flavor, nil
		}
	}

	// fetch by name
	name := flavorName(instance)
	return findFlavorByName(name, compute)
}

func findFlavorByName(name string, compute *gophercloud.ServiceClient) (*flavors.Flavor, error) {
	pages, err := flavors.ListDetail(compute, flavors.ListOpts{}).AllPages()
	if err != nil {
		return nil, err
	}

	current, err := flavors.ExtractFlavors(pages)
	if err != nil {
		return nil, err
	}

	for _, flavor := range current {
		if flavor.Name == name {
			return &flavor, nil
		}
	}

	return nil, nil
}

func reconcileFlavor(ctx context.Context, c client.Client, instance *openstackv1beta1.NovaFlavor, flavor *flavors.Flavor, compute *gophercloud.ServiceClient, log logr.Logger) error {
	name := flavorName(instance)

	var err error

	// create new flavor
	if flavor == nil {
		log.Info("Creating flavor", "name", name)
		flavor, err = flavors.Create(compute, flavors.CreateOpts{
			Name:      name,
			RAM:       instance.Spec.RAM,
			VCPUs:     instance.Spec.VCPUs,
			Disk:      instance.Spec.Disk,
			Swap:      instance.Spec.Swap,
			Ephemeral: instance.Spec.Ephemeral,
			IsPublic:  instance.Spec.IsPublic,
		}).Extract()
		if err != nil {
			return fmt.Errorf("creating flavor: %w", err)
		}
	}

	// TODO replace flavor if spec changed

	// ensure flavor ID present in status
	if instance.Status.FlavorID != flavor.ID {
		instance.Status.FlavorID = flavor.ID
		if err := c.Status().Update(ctx, instance); err != nil {
			return err
		}
	}

	return nil
}

func Delete(instance *openstackv1beta1.NovaFlavor, compute *gophercloud.ServiceClient, log logr.Logger) error {
	flavor, err := getFlavor(instance, compute)
	if err != nil {
		return err
	} else if flavor == nil {
		log.Info("Flavor not found for deletion", "name", instance.Name)
		return nil
	}

	log.Info("Deleting flavor", "name", instance.Name)
	if err := flavors.Delete(compute, flavor.ID).Err; err != nil {
		if _, ok := err.(gophercloud.ErrDefault404); !ok {
			return err
		}
		log.Info("Flavor not found on deletion", "name", instance.Name)
	}

	return nil
}

func flavorName(instance *openstackv1beta1.NovaFlavor) string {
	if instance.Spec.Name == "" {
		return instance.Name
	}
	return instance.Spec.Name
}

func Ensure(ctx context.Context, c client.Client, instance *openstackv1beta1.NovaFlavor, log logr.Logger) error {
	hash, err := template.ObjectHash(instance)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}
	intended := instance.DeepCopy()

	if err := c.Get(ctx, client.ObjectKeyFromObject(instance), instance); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}

		template.SetAppliedHash(instance, hash)

		log.Info("Creating NovaFlavor", "Name", instance.Name)
		return c.Create(ctx, instance)
	} else if !template.MatchesAppliedHash(instance, hash) {
		instance.Spec = intended.Spec

		template.SetAppliedHash(instance, hash)

		log.Info("Updating NovaFlavor", "Name", instance.Name)
		return c.Update(ctx, instance)
	}

	return nil
}

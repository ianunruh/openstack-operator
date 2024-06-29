package service

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/endpoints"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/services"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const defaultRegion = "RegionOne"

func Reconcile(instance *openstackv1beta1.KeystoneService, identity *gophercloud.ServiceClient, log logr.Logger) error {
	svc, err := getService(instance, identity)
	if err != nil {
		return err
	}

	if err := reconcileService(instance, svc, identity, log); err != nil {
		return err
	}

	return nil
}

func getService(instance *openstackv1beta1.KeystoneService, identity *gophercloud.ServiceClient) (*services.Service, error) {
	pages, err := services.List(identity, services.ListOpts{
		Name:        instance.Spec.Name,
		ServiceType: instance.Spec.Type,
	}).AllPages()
	if err != nil {
		return nil, fmt.Errorf("listing services: %w", err)
	}

	current, err := services.ExtractServices(pages)
	if err != nil {
		return nil, fmt.Errorf("extracting services: %w", err)
	}

	if len(current) == 0 {
		return nil, nil
	}

	return &current[0], nil
}

func reconcileService(instance *openstackv1beta1.KeystoneService, svc *services.Service, identity *gophercloud.ServiceClient, log logr.Logger) error {
	var err error

	if svc == nil {
		log.Info("Creating service", "name", instance.Spec.Name)
		svc, err = services.Create(identity, services.CreateOpts{
			Type: instance.Spec.Type,
			Extra: map[string]interface{}{
				"name": instance.Spec.Name,
			},
		}).Extract()
		if err != nil {
			return err
		}
	}

	if err := reconcileEndpoints(instance, svc, identity, log); err != nil {
		return err
	}

	return nil
}

func reconcileEndpoints(instance *openstackv1beta1.KeystoneService, svc *services.Service, identity *gophercloud.ServiceClient, log logr.Logger) error {
	pages, err := endpoints.List(identity, endpoints.ListOpts{
		ServiceID: svc.ID,
	}).AllPages()
	if err != nil {
		return fmt.Errorf("listing endpoints: %w", err)
	}

	current, err := endpoints.ExtractEndpoints(pages)
	if err != nil {
		return fmt.Errorf("extracting endpoints: %w", err)
	}

	expected := map[gophercloud.Availability]string{
		gophercloud.AvailabilityAdmin:    instance.Spec.PublicURL,
		gophercloud.AvailabilityInternal: instance.Spec.InternalURL,
		gophercloud.AvailabilityPublic:   instance.Spec.PublicURL,
	}
	for endpointType, endpointURL := range expected {
		if err := reconcileEndpoint(defaultRegion, endpointType, endpointURL, svc, current, identity, log); err != nil {
			return err
		}
	}

	return nil
}

func reconcileEndpoint(region string, endpointType gophercloud.Availability, endpointURL string, svc *services.Service, current []endpoints.Endpoint, identity *gophercloud.ServiceClient, log logr.Logger) error {
	log = log.WithValues(
		"service", svc.Type,
		"region", region,
		"type", string(endpointType),
		"url", endpointURL)

	endpoint := filterEndpoints(current, region, endpointType)
	if endpoint != nil {
		if endpoint.URL == endpointURL {
			// endpoint unchanged, skip update
			return nil
		}

		log.Info("Updating endpoint", "id", endpoint.ID)
		_, err := endpoints.Update(identity, endpoint.ID, endpoints.UpdateOpts{
			URL: endpointURL,
		}).Extract()
		return err
	}

	log.Info("Creating endpoint")
	_, err := endpoints.Create(identity, endpoints.CreateOpts{
		Availability: endpointType,
		Name:         string(endpointType),
		Region:       region,
		ServiceID:    svc.ID,
		URL:          endpointURL,
	}).Extract()
	return err
}

func filterEndpoints(current []endpoints.Endpoint, region string, availability gophercloud.Availability) *endpoints.Endpoint {
	for _, endpoint := range current {
		if endpoint.Region == region && endpoint.Availability == availability {
			return &endpoint
		}
	}

	return nil
}

func Delete(instance *openstackv1beta1.KeystoneService, identity *gophercloud.ServiceClient, log logr.Logger) error {
	svc, err := getService(instance, identity)
	if err != nil {
		return err
	} else if svc == nil {
		log.Info("Service not found for deletion", "name", instance.Name)
		return nil
	}

	log.Info("Deleting service", "name", instance.Name)
	if err := services.Delete(identity, svc.ID).Err; err != nil {
		if _, ok := err.(gophercloud.ErrDefault404); !ok {
			return err
		}
		log.Info("Service not found on deletion", "name", instance.Name)
	}

	return nil
}

func Ensure(ctx context.Context, c client.Client, instance *openstackv1beta1.KeystoneService, log logr.Logger) error {
	hash, err := template.ObjectHash(instance)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}
	intended := instance.DeepCopy()

	if err := c.Get(ctx, client.ObjectKeyFromObject(instance), instance); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		template.SetAppliedHash(instance, hash)

		log.Info("Creating KeystoneService", "Name", instance.Name)
		return c.Create(ctx, instance)
	} else if !template.MatchesAppliedHash(instance, hash) {
		instance.Spec = intended.Spec

		template.SetAppliedHash(instance, hash)

		log.Info("Updating KeystoneService", "Name", instance.Name)
		return c.Update(ctx, instance)
	}

	return nil
}

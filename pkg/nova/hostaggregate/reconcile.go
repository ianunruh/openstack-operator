package hostaggregate

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/aggregates"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func Reconcile(ctx context.Context, c client.Client, instance *openstackv1beta1.NovaHostAggregate, compute *gophercloud.ServiceClient, log logr.Logger) error {
	aggregate, err := getHostAggregate(instance, compute)
	if err != nil {
		return err
	}

	if err := reconcileHostAggregate(ctx, c, instance, aggregate, compute, log); err != nil {
		return err
	}

	return nil
}

func getHostAggregate(instance *openstackv1beta1.NovaHostAggregate, compute *gophercloud.ServiceClient) (*aggregates.Aggregate, error) {
	// fetch by ID
	if instance.Status.AggregateID > 0 {
		aggregate, err := aggregates.Get(compute, instance.Status.AggregateID).Extract()
		if err != nil {
			if !errors.Is(err, gophercloud.ErrDefault404{}) {
				return nil, err
			}
		} else if aggregate != nil {
			return aggregate, nil
		}
	}

	// fetch by name
	return findHostAggregateByName(instance.Name, compute)
}

func reconcileHostAggregate(ctx context.Context, c client.Client, instance *openstackv1beta1.NovaHostAggregate, aggregate *aggregates.Aggregate, compute *gophercloud.ServiceClient, log logr.Logger) error {
	var err error

	// create new aggregate
	if aggregate == nil {
		log.Info("Creating host aggregate", "name", instance.Name)
		aggregate, err = aggregates.Create(compute, aggregates.CreateOpts{
			Name:             instance.Name,
			AvailabilityZone: instance.Spec.Zone,
		}).Extract()
		if err != nil {
			return err
		}
	}

	// ensure aggregate ID present in status
	if instance.Status.AggregateID == 0 {
		instance.Status.AggregateID = aggregate.ID
		if err := c.Status().Update(ctx, instance); err != nil {
			return err
		}
	}

	// reconcile AZ
	if aggregate.AvailabilityZone != instance.Spec.Zone {
		log.Info("Updating host aggregate availability zone", "name", instance.Name)
		aggregate, err = aggregates.Update(compute, aggregate.ID, aggregates.UpdateOpts{
			AvailabilityZone: instance.Spec.Zone,
		}).Extract()
		if err != nil {
			return err
		}
	}

	// reconcile metadata
	metadata := metadataWithDefaults(instance.Spec.Metadata, instance.Spec.Zone)
	if metadataChanged(metadata, aggregate.Metadata) {
		log.Info("Updating host aggregate metadata", "name", instance.Name)
		aggregate, err = aggregates.SetMetadata(compute, aggregate.ID, aggregates.SetMetadataOpts{
			Metadata: castMetadataValuesToInterface(metadata),
		}).Extract()
		if err != nil {
			return err
		}
	}

	// reconcile hosts, if node selector is present
	if len(instance.Spec.NodeSelector) > 0 {
		var nodes corev1.NodeList
		if err := c.List(ctx, &nodes, &client.ListOptions{
			LabelSelector: labels.Set(instance.Spec.NodeSelector).AsSelector(),
		}); err != nil {
			return err
		}

		currentHosts := make(map[string]bool, len(aggregate.Hosts))
		for _, host := range aggregate.Hosts {
			currentHosts[host] = true
		}

		// add new nodes to aggregate
		nodesByHost := make(map[string]corev1.Node, len(nodes.Items))
		for _, node := range nodes.Items {
			// TODO depending on cluster setup, this may need to be an FQDN
			host := node.Name
			nodesByHost[host] = node

			if currentHosts[host] {
				continue
			}

			log.Info("Adding node to host aggregate", "name", instance.Name, "host", host)
			aggregate, err = aggregates.AddHost(compute, aggregate.ID, aggregates.AddHostOpts{
				Host: host,
			}).Extract()
			if err != nil {
				return err
			}
		}

		// remove deleted nodes from aggregate
		for host := range currentHosts {
			if _, ok := nodesByHost[host]; ok {
				continue
			}

			log.Info("Removing node from host aggregate", "name", instance.Name, "host", host)
			aggregate, err = aggregates.RemoveHost(compute, aggregate.ID, aggregates.RemoveHostOpts{
				Host: host,
			}).Extract()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func findHostAggregateByName(name string, compute *gophercloud.ServiceClient) (*aggregates.Aggregate, error) {
	pages, err := aggregates.List(compute).AllPages()
	if err != nil {
		return nil, err
	}

	current, err := aggregates.ExtractAggregates(pages)
	if err != nil {
		return nil, err
	}

	for _, aggregate := range current {
		if aggregate.Name == name {
			return &aggregate, nil
		}
	}

	return nil, nil
}

func Delete(instance *openstackv1beta1.NovaHostAggregate, compute *gophercloud.ServiceClient, log logr.Logger) error {
	aggregate, err := getHostAggregate(instance, compute)
	if err != nil {
		return err
	} else if aggregate == nil {
		log.Info("Aggregate not found for deletion", "name", instance.Name)
		return nil
	}

	for _, host := range aggregate.Hosts {
		log.Info("Removing node from host aggregate", "name", instance.Name, "host", host)
		if err := aggregates.RemoveHost(compute, aggregate.ID, aggregates.RemoveHostOpts{
			Host: host,
		}).Err; err != nil {
			if errors.Is(err, gophercloud.ErrDefault404{}) {
				log.Info("Aggregate host not found for deletion", "name", instance.Name, "host", host)
			} else {
				return err
			}
		}
	}

	log.Info("Deleting host aggregate", "name", instance.Name)
	if err := aggregates.Delete(compute, aggregate.ID).Err; err != nil {
		if errors.Is(err, gophercloud.ErrDefault404{}) {
			log.Info("Aggregate not found on deletion", "name", instance.Name)
		} else {
			return err
		}
	}

	return nil
}

func metadataWithDefaults(metadata map[string]string, az string) map[string]string {
	out := make(map[string]string, len(metadata))
	for key, value := range metadata {
		out[key] = value
	}
	if az != "" {
		out["availability_zone"] = az
	}
	return out
}

func metadataChanged(expected, current map[string]string) bool {
	if len(expected) != len(current) {
		return true
	}

	for key, value := range expected {
		if current[key] != value {
			return true
		}
	}

	return false
}

func castMetadataValuesToInterface(metadata map[string]string) map[string]interface{} {
	out := make(map[string]interface{}, len(metadata))
	for key, value := range metadata {
		out[key] = value
	}
	return out
}

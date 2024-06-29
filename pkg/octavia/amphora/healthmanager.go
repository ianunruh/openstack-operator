package amphora

import (
	"context"
	"sort"

	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	"golang.org/x/exp/maps"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func (b *bootstrap) EnsureHealthPorts(ctx context.Context) error {
	status := b.instance.Status.Amphora

	networkID := status.NetworkIDs[0]
	securityGroups := status.HealthSecurityGroupIDs

	nodes, err := b.listHealthMangerNodes(ctx)
	if err != nil {
		return err
	}

	currentByName, err := b.listHealthManagerPorts()
	if err != nil {
		return err
	}

	portsByName := make(map[string]openstackv1beta1.OctaviaAmphoraHealthPort)
	for _, node := range nodes {
		name := healthPortPrefix + node.Name

		port, ok := currentByName[name]
		if !ok {
			b.log.Info("Creating port",
				"name", name,
				"networkID", networkID)
			result, err := ports.Create(b.network, ports.CreateOpts{
				Name:           name,
				NetworkID:      networkID,
				SecurityGroups: &securityGroups,
				DeviceOwner:    healthPortDeviceOwner,
			}).Extract()
			if err != nil {
				return err
			}
			port = *result
		}

		portsByName[name] = openstackv1beta1.OctaviaAmphoraHealthPort{
			ID:         port.ID,
			Name:       port.Name,
			MACAddress: port.MACAddress,
			IPAddress:  port.FixedIPs[0].IPAddress,
		}
	}

	for name, port := range currentByName {
		if _, ok := portsByName[name]; ok {
			continue
		}

		b.log.Info("Deleting port",
			"name", name,
			"networkID", networkID)
		if err := ports.Delete(b.network, port.ID).ExtractErr(); err != nil {
			return err
		}
	}

	healthPorts := maps.Values(portsByName)
	sortPortsByName(healthPorts)

	if !portsChanged(healthPorts, b.instance.Status.Amphora.HealthPorts) {
		return nil
	}

	b.instance.Status.Amphora.HealthPorts = healthPorts
	if err := b.client.Status().Update(ctx, b.instance); err != nil {
		return err
	}

	return nil
}

func (b *bootstrap) listHealthMangerNodes(ctx context.Context) ([]corev1.Node, error) {
	nodeLabelSelector, err := labels.ValidatedSelectorFromSet(b.instance.Spec.HealthManager.NodeSelector)
	if err != nil {
		return nil, err
	}

	nodes := &corev1.NodeList{}
	if err := b.client.List(ctx, nodes, &client.ListOptions{
		LabelSelector: nodeLabelSelector,
	}); err != nil {
		return nil, err
	}

	return nodes.Items, nil
}

func (b *bootstrap) listHealthManagerPorts() (map[string]ports.Port, error) {
	page, err := ports.List(b.network, ports.ListOpts{
		NetworkID:   b.instance.Status.Amphora.NetworkIDs[0],
		DeviceOwner: healthPortDeviceOwner,
	}).AllPages()
	if err != nil {
		return nil, err
	}

	result, err := ports.ExtractPorts(page)
	if err != nil {
		return nil, err
	}

	byName := make(map[string]ports.Port)
	for _, port := range result {
		byName[port.Name] = port
	}
	return byName, nil
}

func sortPortsByName(ports []openstackv1beta1.OctaviaAmphoraHealthPort) {
	sort.Slice(ports, func(i, j int) bool {
		return ports[i].Name < ports[j].Name
	})
}

func portsChanged(left, right []openstackv1beta1.OctaviaAmphoraHealthPort) bool {
	if len(left) != len(right) {
		return true
	}
	for i, port := range left {
		if port != right[i] {
			return true
		}
	}
	return false
}

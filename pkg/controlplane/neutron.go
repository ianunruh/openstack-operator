package controlplane

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func Neutron(instance *openstackv1beta1.ControlPlane, nova *openstackv1beta1.Nova) *openstackv1beta1.Neutron {
	// TODO labels
	spec := instance.Spec.Neutron

	spec.Server.Ingress = ingressDefaults(spec.Server.Ingress, instance, "neutron")
	spec.Server.NodeSelector = controllerNodeSelector(spec.Server.NodeSelector, instance)
	// spec.Server.TLS = tlsServerDefaults(spec.Server.TLS, instance)

	spec.DBSyncJob.NodeSelector = controllerNodeSelector(spec.DBSyncJob.NodeSelector, instance)

	spec.MetadataAgent.NodeSelector = computeNodeSelector(spec.MetadataAgent.NodeSelector, instance)

	spec.Broker = brokerUserDefaults(spec.Broker, instance)
	spec.Cache = cacheDefaults(spec.Cache, instance)
	spec.Database = databaseDefaults(spec.Database, instance)
	spec.TLS = tlsClientDefaults(spec.TLS, instance)

	spec.Nova = neutronNovaDefaults(spec.Nova, nova)

	return &openstackv1beta1.Neutron{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "neutron",
			Namespace: instance.Namespace,
		},
		Spec: spec,
	}
}

func neutronNovaDefaults(spec openstackv1beta1.NeutronNovaSpec, nova *openstackv1beta1.Nova) openstackv1beta1.NeutronNovaSpec {
	// TODO support multiple cells
	cell := nova.Spec.Cells[0]

	if spec.MetadataHost == "" {
		spec.MetadataHost = template.Combine(nova.Name, cell.Name, "metadata")
	}

	if spec.MetadataProtocol == "" && isTLSEnabled(cell.Metadata.TLS) {
		spec.MetadataProtocol = "https"
	}

	return spec
}

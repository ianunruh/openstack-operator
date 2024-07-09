package controlplane

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

const (
	DefaultGlancePVCSize = "100Gi"
)

func Glance(instance *openstackv1beta1.ControlPlane) *openstackv1beta1.Glance {
	spec := instance.Spec.Glance

	spec.API.Ingress = ingressDefaults(spec.API.Ingress, instance, "glance")
	spec.API.NodeSelector = controllerNodeSelector(spec.API.NodeSelector, instance)
	spec.API.TLS = tlsServerDefaults(spec.API.TLS, instance)

	spec.DBSyncJob.NodeSelector = controllerNodeSelector(spec.DBSyncJob.NodeSelector, instance)

	spec.Cache = cacheDefaults(spec.Cache, instance)
	spec.Database = databaseDefaults(spec.Database, instance)
	spec.TLS = tlsClientDefaults(spec.TLS, instance)

	if spec.Backends == nil {
		spec.Backends = []openstackv1beta1.GlanceBackendSpec{
			{
				Name: "ssd",
				PVC: &openstackv1beta1.VolumeSpec{
					Capacity: DefaultGlancePVCSize,
				},
			},
		}
	}

	return &openstackv1beta1.Glance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "glance",
			Namespace: instance.Namespace,
		},
		Spec: spec,
	}
}

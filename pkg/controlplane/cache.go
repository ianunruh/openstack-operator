package controlplane

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func Cache(instance *openstackv1beta1.ControlPlane) *openstackv1beta1.Memcached {
	if instance.Spec.ExternalCache != nil {
		return nil
	}

	spec := instance.Spec.Cache

	spec.NodeSelector = controllerNodeSelector(spec.NodeSelector, instance)

	// TODO labels
	return &openstackv1beta1.Memcached{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "memcached",
			Namespace: instance.Namespace,
		},
		Spec: spec,
	}
}

func cacheDefaults(spec openstackv1beta1.CacheSpec, instance *openstackv1beta1.ControlPlane) openstackv1beta1.CacheSpec {
	if spec.IsZero() {
		return *instance.Spec.ExternalCache
	}
	return spec
}

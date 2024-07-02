package barbican

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func KeystoneService(instance *openstackv1beta1.Barbican) *openstackv1beta1.KeystoneService {
	labels := template.AppLabels(instance.Name, AppLabel)

	return &openstackv1beta1.KeystoneService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: openstackv1beta1.KeystoneServiceSpec{
			Name:        "barbican",
			Type:        "key-manager",
			InternalURL: APIInternalURL(instance),
			PublicURL:   APIPublicURL(instance),
		},
	}
}

func KeystoneUser(instance *openstackv1beta1.Barbican) *openstackv1beta1.KeystoneUser {
	labels := template.AppLabels(instance.Name, AppLabel)

	return &openstackv1beta1.KeystoneUser{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: openstackv1beta1.KeystoneUserSpec{
			Secret:  template.Combine(instance.Name, "keystone"),
			Project: "service",
		},
	}
}

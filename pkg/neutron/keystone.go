package neutron

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func KeystoneService(instance *openstackv1beta1.Neutron) *openstackv1beta1.KeystoneService {
	labels := template.AppLabels(instance.Name, AppLabel)

	return &openstackv1beta1.KeystoneService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: openstackv1beta1.KeystoneServiceSpec{
			Name:        "neutron",
			Type:        "network",
			InternalURL: fmt.Sprintf("http://%s-server.%s.svc:9696", instance.Name, instance.Namespace),
			PublicURL:   fmt.Sprintf("https://%s", instance.Spec.Server.Ingress.Host),
		},
	}
}

func KeystoneUser(instance *openstackv1beta1.Neutron) *openstackv1beta1.KeystoneUser {
	labels := template.AppLabels(instance.Name, AppLabel)

	return &openstackv1beta1.KeystoneUser{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: openstackv1beta1.KeystoneUserSpec{
			Secret: template.Combine(instance.Name, "keystone"),
		},
	}
}

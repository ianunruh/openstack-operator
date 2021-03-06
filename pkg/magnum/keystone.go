package magnum

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func KeystoneService(instance *openstackv1beta1.Magnum) *openstackv1beta1.KeystoneService {
	labels := template.AppLabels(instance.Name, AppLabel)

	return &openstackv1beta1.KeystoneService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: openstackv1beta1.KeystoneServiceSpec{
			Name:        "magnum",
			Type:        "container-infra",
			InternalURL: fmt.Sprintf("http://%s-api.%s.svc:9511/v1", instance.Name, instance.Namespace),
			PublicURL:   fmt.Sprintf("https://%s/v1", instance.Spec.API.Ingress.Host),
		},
	}
}

func KeystoneUser(instance *openstackv1beta1.Magnum) *openstackv1beta1.KeystoneUser {
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

func KeystoneStackUser(instance *openstackv1beta1.Magnum) *openstackv1beta1.KeystoneUser {
	labels := template.AppLabels(instance.Name, AppLabel)

	name := template.Combine(instance.Name, "stack")

	return &openstackv1beta1.KeystoneUser{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: openstackv1beta1.KeystoneUserSpec{
			Secret: template.Combine(name, "keystone"),
			Domain: "magnum",
		},
	}
}

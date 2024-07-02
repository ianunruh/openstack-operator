package heat

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func KeystoneServices(instance *openstackv1beta1.Heat) []*openstackv1beta1.KeystoneService {
	labels := template.AppLabels(instance.Name, AppLabel)

	return []*openstackv1beta1.KeystoneService{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      instance.Name,
				Namespace: instance.Namespace,
				Labels:    labels,
			},
			Spec: openstackv1beta1.KeystoneServiceSpec{
				Name:        "heat",
				Type:        "orchestration",
				InternalURL: APIInternalURL(instance),
				PublicURL:   APIPublicURL(instance),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      template.Combine(instance.Name, "cfn"),
				Namespace: instance.Namespace,
				Labels:    labels,
			},
			Spec: openstackv1beta1.KeystoneServiceSpec{
				Name:        "heat-cfn",
				Type:        "cloudformation",
				InternalURL: CFNInternalURL(instance),
				PublicURL:   CFNPublicURL(instance),
			},
		},
	}
}

func KeystoneUser(instance *openstackv1beta1.Heat) *openstackv1beta1.KeystoneUser {
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

func KeystoneStackUser(instance *openstackv1beta1.Heat) *openstackv1beta1.KeystoneUser {
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
			Domain: "heat",
		},
	}
}

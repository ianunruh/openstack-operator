package manila

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func KeystoneServices(instance *openstackv1beta1.Manila) []*openstackv1beta1.KeystoneService {
	labels := template.AppLabels(instance.Name, AppLabel)

	return []*openstackv1beta1.KeystoneService{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      instance.Name,
				Namespace: instance.Namespace,
				Labels:    labels,
			},
			Spec: openstackv1beta1.KeystoneServiceSpec{
				Name:        "manila",
				Type:        "share",
				InternalURL: fmt.Sprintf("http://%s-api.%s.svc:8786/v1/%%(tenant_id)s", instance.Name, instance.Namespace),
				PublicURL:   fmt.Sprintf("https://%s/v1/%%(tenant_id)s", instance.Spec.API.Ingress.Host),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      template.Combine(instance.Name, "v2"),
				Namespace: instance.Namespace,
				Labels:    labels,
			},
			Spec: openstackv1beta1.KeystoneServiceSpec{
				Name:        "manilav2",
				Type:        "sharev2",
				InternalURL: fmt.Sprintf("http://%s-api.%s.svc:8786/v2", instance.Name, instance.Namespace),
				PublicURL:   fmt.Sprintf("https://%s/v2", instance.Spec.API.Ingress.Host),
			},
		},
	}
}

func KeystoneUser(instance *openstackv1beta1.Manila) *openstackv1beta1.KeystoneUser {
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

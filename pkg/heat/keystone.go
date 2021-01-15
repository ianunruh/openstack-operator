package heat

import (
	"fmt"

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
				InternalURL: fmt.Sprintf("http://%s-api.%s.svc:8004/v1/$(project_id)s", instance.Name, instance.Namespace),
				PublicURL:   fmt.Sprintf("https://%s/v1/$(project_id)s", instance.Spec.API.Ingress.Host),
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
				InternalURL: fmt.Sprintf("http://%s-cfn.%s.svc:8000/v1", instance.Name, instance.Namespace),
				PublicURL:   fmt.Sprintf("https://%s/v1", instance.Spec.CFN.Ingress.Host),
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
			Secret: template.Combine(instance.Name, "keystone"),
		},
	}
}

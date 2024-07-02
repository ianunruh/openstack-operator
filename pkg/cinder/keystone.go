package cinder

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func KeystoneServices(instance *openstackv1beta1.Cinder) []*openstackv1beta1.KeystoneService {
	labels := template.AppLabels(instance.Name, AppLabel)

	return []*openstackv1beta1.KeystoneService{
		{
			ObjectMeta: metav1.ObjectMeta{
				// TODO it appears that v2 is deprecated now and this can probs be removed
				Name:      template.Combine(instance.Name, "v2"),
				Namespace: instance.Namespace,
				Labels:    labels,
			},
			Spec: openstackv1beta1.KeystoneServiceSpec{
				Name:        "cinderv2",
				Type:        "volumev2",
				InternalURL: APIInternalURL(instance, "v2"),
				PublicURL:   APIPublicURL(instance, "v2"),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      template.Combine(instance.Name, "v3"),
				Namespace: instance.Namespace,
				Labels:    labels,
			},
			Spec: openstackv1beta1.KeystoneServiceSpec{
				Name:        "cinderv3",
				Type:        "volumev3",
				InternalURL: APIInternalURL(instance, "v3"),
				PublicURL:   APIPublicURL(instance, "v3"),
			},
		},
	}
}

func KeystoneUser(instance *openstackv1beta1.Cinder) *openstackv1beta1.KeystoneUser {
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

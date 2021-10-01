package nova

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func Cell(instance *openstackv1beta1.Nova, spec openstackv1beta1.NovaCellSpec) *openstackv1beta1.NovaCell {
	labels := template.AppLabels(instance.Name, AppLabel)

	return &openstackv1beta1.NovaCell{
		ObjectMeta: metav1.ObjectMeta{
			Name:      template.Combine(instance.Name, spec.Name),
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: spec,
	}
}

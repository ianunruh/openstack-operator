package controlplane

import (
	"fmt"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func ingressDefaults(spec *openstackv1beta1.IngressSpec, instance *openstackv1beta1.ControlPlane, name string) *openstackv1beta1.IngressSpec {
	if spec == nil {
		spec = &openstackv1beta1.IngressSpec{}
	}

	common := instance.Spec.Ingress

	spec.Annotations = template.MergeStringMaps(common.Annotations, spec.Annotations)

	if spec.Host == "" {
		spec.Host = fmt.Sprintf("%s.%s", name, instance.Spec.Domain)
	}

	if spec.ClassName == nil {
		spec.ClassName = common.ClassName
	}

	if spec.TLSSecretName == "" {
		spec.TLSSecretName = common.TLSSecretName
	}

	return spec
}

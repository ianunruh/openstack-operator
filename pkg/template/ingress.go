package template

import (
	"context"

	"github.com/go-logr/logr"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func EnsureIngress(ctx context.Context, c client.Client, instance *netv1.Ingress, log logr.Logger) error {
	return Ensure(ctx, c, instance, log, func(intended *netv1.Ingress) {
		instance.Spec = intended.Spec
		instance.Annotations = intended.Annotations
	})
}

func IngressServiceBackend(svcName, portName string) netv1.IngressBackend {
	return netv1.IngressBackend{
		Service: &netv1.IngressServiceBackend{
			Name: svcName,
			Port: netv1.ServiceBackendPort{
				Name: portName,
			},
		},
	}
}

func GenericIngress(name, namespace string, spec *openstackv1beta1.IngressSpec, labels map[string]string) *netv1.Ingress {
	prefixPathType := netv1.PathTypePrefix

	tlsSecretName := spec.TLSSecretName
	if tlsSecretName == "" {
		tlsSecretName = Combine(name, "ingress-tls")
	}

	ingress := &netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: spec.Annotations,
		},
		Spec: netv1.IngressSpec{
			IngressClassName: spec.ClassName,
			TLS: []netv1.IngressTLS{
				{
					SecretName: tlsSecretName,
					Hosts:      []string{spec.Host},
				},
			},
			Rules: []netv1.IngressRule{
				{
					Host: spec.Host,
					IngressRuleValue: netv1.IngressRuleValue{
						HTTP: &netv1.HTTPIngressRuleValue{
							Paths: []netv1.HTTPIngressPath{
								{
									PathType: &prefixPathType,
									Path:     "/",
									Backend:  IngressServiceBackend(name, "http"),
								},
							},
						},
					},
				},
			},
		},
	}

	return ingress
}

func GenericIngressWithTLS(name, namespace string, spec *openstackv1beta1.IngressSpec, tlsSpec openstackv1beta1.TLSServerSpec, labels map[string]string) *netv1.Ingress {
	ingress := GenericIngress(name, namespace, spec, labels)

	if tlsSpec.Secret != "" {
		ingress.Annotations["nginx.ingress.kubernetes.io/backend-protocol"] = "HTTPS"
	}

	return ingress
}

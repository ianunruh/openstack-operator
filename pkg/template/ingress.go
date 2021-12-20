package template

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

func EnsureIngress(ctx context.Context, c client.Client, intended *netv1.Ingress, log logr.Logger) error {
	hash, err := ObjectHash(intended)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}

	found := &netv1.Ingress{}
	if err := c.Get(ctx, client.ObjectKeyFromObject(intended), found); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		SetAppliedHash(intended, hash)

		log.Info("Creating Ingress", "Name", intended.Name)
		return c.Create(ctx, intended)
	} else if !MatchesAppliedHash(found, hash) {
		found.Spec = intended.Spec
		found.Annotations = intended.Annotations
		SetAppliedHash(found, hash)

		log.Info("Updating Ingress", "Name", intended.Name)
		return c.Update(ctx, found)
	}

	return nil
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

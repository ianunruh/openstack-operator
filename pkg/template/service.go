package template

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func HeadlessServiceName(name string) string {
	return Combine(name, "headless")
}

func GenericService(name, namespace string, labels map[string]string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
		},
	}
}

func EnsureService(ctx context.Context, c client.Client, instance *corev1.Service, log logr.Logger) error {
	return Ensure(ctx, c, instance, log, func(intended *corev1.Service) {
		// copy immutable fields
		intended.Spec.ClusterIP = instance.Spec.ClusterIP

		instance.Spec = intended.Spec
	})
}

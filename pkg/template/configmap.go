package template

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GenericConfigMap(name, namespace string, labels map[string]string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Data: map[string]string{},
	}
}

func EnsureConfigMap(ctx context.Context, c client.Client, instance *corev1.ConfigMap, log logr.Logger) error {
	return Ensure(ctx, c, instance, log, func(intended *corev1.ConfigMap) {
		instance.Data = intended.Data
	})
}

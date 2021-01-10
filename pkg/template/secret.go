package template

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GenericSecret(name, namespace string, labels map[string]string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		StringData: map[string]string{},
	}
}

func NewPassword() string {
	return utilrand.String(20)
}

func CreateSecret(ctx context.Context, c client.Client, intended *corev1.Secret, log logr.Logger) (*corev1.Secret, error) {
	found := &corev1.Secret{}
	err := c.Get(ctx, client.ObjectKeyFromObject(intended), found)
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		}

		log.Info("Creating Secret", "Name", intended.Name)
		return intended, c.Create(ctx, intended)
	}

	return found, nil
}

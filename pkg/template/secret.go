package template

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/rand"

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

func NewFernetKey() string {
	data := make([]byte, 32)
	for i := 0; i < 32; i++ {
		data[i] = byte(rand.Intn(10))
	}
	return base64.StdEncoding.EncodeToString(data)
}

func CreateSecret(ctx context.Context, c client.Client, instance *corev1.Secret, log logr.Logger) error {
	err := c.Get(ctx, client.ObjectKeyFromObject(instance), instance)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		log.Info("Creating Secret", "Name", instance.Name)
		return c.Create(ctx, instance)
	}

	return nil
}

func EnsureSecret(ctx context.Context, c client.Client, intended *corev1.Secret, log logr.Logger) error {
	hash, err := ObjectHash(intended)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}
	SetAppliedHash(intended, hash)

	found := &corev1.Secret{}
	if err := c.Get(ctx, client.ObjectKeyFromObject(intended), found); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		log.Info("Creating Secret", "Name", intended.Name)
		return c.Create(ctx, intended)
	} else if !MatchesAppliedHash(found, hash) {
		found.Data = intended.Data
		SetAppliedHash(found, hash)

		log.Info("Updating Secret", "Name", intended.Name)
		return c.Update(ctx, found)
	}

	return nil
}

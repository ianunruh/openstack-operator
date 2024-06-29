package amphora

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	novakeypair "github.com/ianunruh/openstack-operator/pkg/nova/keypair"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func (b *bootstrap) EnsureKeypair(ctx context.Context) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      template.Combine(b.instance.Name, "amphora-ssh"),
			Namespace: b.instance.Namespace,
		},
	}
	if err := b.client.Get(ctx, client.ObjectKeyFromObject(secret), secret); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		secret, err = newKeypairSecret(b.instance)
		if err != nil {
			return err
		}
		b.log.Info("Creating keypair secret", "name", secret.Name)
		if err := b.client.Create(ctx, secret); err != nil {
			return err
		}
	}

	keypair := &openstackv1beta1.NovaKeypair{
		ObjectMeta: metav1.ObjectMeta{
			Name:      template.Combine(b.instance.Name, "amphora"),
			Namespace: b.instance.Namespace,
		},
		Spec: openstackv1beta1.NovaKeypairSpec{
			Name:      amphoraKeypairName,
			PublicKey: string(secret.Data["id_rsa.pub"]),
			User:      b.instance.Name,
		},
	}
	controllerutil.SetControllerReference(b.instance, keypair, b.client.Scheme())
	if err := novakeypair.Ensure(ctx, b.client, keypair, b.log); err != nil {
		return err
	}

	novakeypair.AddReadyCheck(b.deps, keypair)

	return nil
}

func newKeypairSecret(instance *openstackv1beta1.Octavia) (*corev1.Secret, error) {
	labels := template.AppLabels(instance.Name, "octavia")
	name := template.Combine(instance.Name, "amphora-ssh")

	return template.SSHKeypairSecret(name, instance.Namespace, labels)
}

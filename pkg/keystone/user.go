package keystone

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func UserJob(instance *openstackv1beta1.KeystoneUser, containerImage, adminSecret string) *batchv1.Job {
	labels := template.AppLabels(instance.Name, AppLabel)

	job := template.GenericJob(template.Component{
		Namespace: instance.Namespace,
		Labels:    labels,
		Containers: []corev1.Container{
			{
				Name:  "database",
				Image: containerImage,
				Command: []string{
					"python3",
					"-c",
					template.MustRenderFile(AppLabel, "user-setup.py", nil),
				},
				EnvFrom: []corev1.EnvFromSource{
					template.EnvFromSecret(adminSecret),
				},
				Env: []corev1.EnvVar{},
			},
		},
	})

	job.Name = template.Combine("keystone", "user", instance.Name)

	return job
}

func UserSecret(instance *openstackv1beta1.KeystoneUser) *corev1.Secret {
	labels := template.AppLabels(instance.Name, AppLabel)
	secret := template.GenericSecret(instance.Spec.Secret, instance.Namespace, labels)
	secret.StringData["password"] = template.NewPassword()
	return secret
}

func EnsureUser(ctx context.Context, c client.Client, intended *openstackv1beta1.KeystoneUser, log logr.Logger) error {
	hash, err := template.ObjectHash(intended)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}

	found := &openstackv1beta1.KeystoneUser{}
	if err := c.Get(ctx, client.ObjectKeyFromObject(intended), found); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		template.SetAppliedHash(intended, hash)

		log.Info("Creating KeystoneUser", "Name", intended.Name)
		return c.Create(ctx, intended)
	} else if !template.MatchesAppliedHash(found, hash) {
		found.Spec = intended.Spec
		template.SetAppliedHash(found, hash)

		log.Info("Updating KeystoneUser", "Name", intended.Name)
		return c.Update(ctx, found)
	}

	return nil
}

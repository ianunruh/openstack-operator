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
				Name:  "setup",
				Image: containerImage,
				Command: []string{
					"python3",
					"-c",
					template.MustRenderFile(AppLabel, "user-setup.py", nil),
				},
				EnvFrom: []corev1.EnvFromSource{
					template.EnvFromSecret(adminSecret),
					template.EnvFromSecretPrefixed(instance.Spec.Secret, "SVC_"),
				},
			},
		},
	})

	job.Name = template.Combine("keystone", "user", instance.Name)

	return job
}

func UserSecret(instance *openstackv1beta1.KeystoneUser) *corev1.Secret {
	labels := template.AppLabels(instance.Name, AppLabel)
	secret := template.GenericSecret(instance.Spec.Secret, instance.Namespace, labels)

	secret.StringData = map[string]string{
		"OS_IDENTITY_API_VERSION": "3",
		"OS_AUTH_URL":             "http://keystone-api:5000/v3",
		"OS_REGION_NAME":          "RegionOne",
		"OS_PROJECT_DOMAIN_NAME":  "Default",
		"OS_USER_DOMAIN_NAME":     "Default",
		"OS_PROJECT_NAME":         "service",
		"OS_USERNAME":             instance.Name,
		"OS_PASSWORD":             template.NewPassword(),
	}

	return secret
}

func EnsureUser(ctx context.Context, c client.Client, instance *openstackv1beta1.KeystoneUser, log logr.Logger) error {
	intended := instance.DeepCopy()
	hash, err := template.ObjectHash(intended)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}

	if err := c.Get(ctx, client.ObjectKeyFromObject(instance), instance); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		template.SetAppliedHash(instance, hash)

		log.Info("Creating KeystoneUser", "Name", instance.Name)
		return c.Create(ctx, instance)
	} else if !template.MatchesAppliedHash(instance, hash) {
		instance.Spec = intended.Spec
		template.SetAppliedHash(instance, hash)

		log.Info("Updating KeystoneUser", "Name", instance.Name)
		return c.Update(ctx, instance)
	}

	return nil
}

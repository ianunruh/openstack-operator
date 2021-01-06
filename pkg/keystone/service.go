package keystone

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func ServiceJob(instance *openstackv1beta1.KeystoneService, containerImage, adminSecret string) *batchv1.Job {
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
					template.MustRenderFile(AppLabel, "service-setup.py", nil),
				},
				EnvFrom: []corev1.EnvFromSource{
					template.EnvFromSecret(adminSecret),
				},
				Env: []corev1.EnvVar{
					template.EnvVar("KEYSTONE_SERVICE_NAME", instance.Spec.Name),
					template.EnvVar("KEYSTONE_SERVICE_TYPE", instance.Spec.Type),
					template.EnvVar("KEYSTONE_SERVICE_ENDPOINT_ADMIN", instance.Spec.PublicURL),
					template.EnvVar("KEYSTONE_SERVICE_ENDPOINT_INTERNAL", instance.Spec.InternalURL),
					template.EnvVar("KEYSTONE_SERVICE_ENDPOINT_PUBLIC", instance.Spec.PublicURL),
				},
			},
		},
	})

	job.Name = template.Combine("keystone", "service", instance.Name)

	return job
}

func EnsureService(ctx context.Context, c client.Client, intended *openstackv1beta1.KeystoneService, log logr.Logger) error {
	hash, err := template.ObjectHash(intended)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}

	found := &openstackv1beta1.KeystoneService{}
	if err := c.Get(ctx, client.ObjectKeyFromObject(intended), found); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		template.SetAppliedHash(intended, hash)

		log.Info("Creating KeystoneService", "Name", intended.Name)
		return c.Create(ctx, intended)
	} else if !template.MatchesAppliedHash(found, hash) {
		found.Spec = intended.Spec
		template.SetAppliedHash(found, hash)

		log.Info("Updating KeystoneService", "Name", intended.Name)
		return c.Update(ctx, found)
	}

	return nil
}

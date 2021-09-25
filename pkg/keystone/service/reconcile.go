package service

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/keystone"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func SetupJob(instance *openstackv1beta1.KeystoneService, containerImage, adminSecret string) *batchv1.Job {
	labels := template.AppLabels(instance.Name, keystone.AppLabel)

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
					template.MustReadFile(keystone.AppLabel, "service-setup.py"),
				},
				EnvFrom: []corev1.EnvFromSource{
					template.EnvFromSecret(adminSecret),
				},
				Env: []corev1.EnvVar{
					template.EnvVar("SVC_NAME", instance.Spec.Name),
					template.EnvVar("SVC_TYPE", instance.Spec.Type),
					template.EnvVar("SVC_REGION", "RegionOne"),
					template.EnvVar("SVC_ENDPOINT_ADMIN", instance.Spec.PublicURL),
					template.EnvVar("SVC_ENDPOINT_INTERNAL", instance.Spec.InternalURL),
					template.EnvVar("SVC_ENDPOINT_PUBLIC", instance.Spec.PublicURL),
				},
			},
		},
	})

	job.Name = template.Combine("keystone", "service", instance.Name)

	return job
}

func Ensure(ctx context.Context, c client.Client, instance *openstackv1beta1.KeystoneService, log logr.Logger) error {
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

		log.Info("Creating KeystoneService", "Name", instance.Name)
		return c.Create(ctx, intended)
	} else if !template.MatchesAppliedHash(instance, hash) {
		instance.Spec = intended.Spec
		template.SetAppliedHash(instance, hash)

		log.Info("Updating KeystoneService", "Name", instance.Name)
		return c.Update(ctx, instance)
	}

	return nil
}

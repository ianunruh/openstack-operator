package cell

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/nova"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

var (
	appUID = int64(42436)
)

func DBSyncJob(instance *openstackv1beta1.NovaCell, env []corev1.EnvVar, volumes []corev1.Volume, containerImage string) *batchv1.Job {
	labels := template.AppLabels(instance.Name, nova.AppLabel)

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-nova", "/etc/nova/nova.conf", "nova.conf"),
	}

	job := template.GenericJob(template.Component{
		Namespace: instance.Namespace,
		Labels:    labels,
		Containers: []corev1.Container{
			{
				Name:  "db-sync",
				Image: containerImage,
				Command: []string{
					"bash",
					"-c",
					template.MustReadFile(nova.AppLabel, "cell-db-sync.sh"),
				},
				Env:          env,
				Resources:    instance.Spec.DBSyncJob.Resources,
				VolumeMounts: volumeMounts,
			},
		},
		NodeSelector: instance.Spec.DBSyncJob.NodeSelector,
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser: &appUID,
			FSGroup:   &appUID,
		},
		Volumes: volumes,
	})

	job.Name = template.Combine(instance.Name, "db-sync")

	return job
}

func Ensure(ctx context.Context, c client.Client, instance *openstackv1beta1.NovaCell, log logr.Logger) error {
	intended := instance.DeepCopy()
	hash, err := template.ObjectHash(instance)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}

	if err := c.Get(ctx, client.ObjectKeyFromObject(instance), instance); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		template.SetAppliedHash(intended, hash)

		log.Info("Creating NovaCell", "Name", instance.Name)
		return c.Create(ctx, instance)
	} else if !template.MatchesAppliedHash(instance, hash) {
		instance.Spec = intended.Spec
		template.SetAppliedHash(instance, hash)

		log.Info("Updating NovaCell", "Name", instance.Name)
		return c.Update(ctx, instance)
	}

	return nil
}

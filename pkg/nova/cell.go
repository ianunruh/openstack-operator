package nova

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

func Cell(instance *openstackv1beta1.Nova, spec openstackv1beta1.NovaCellSpec) *openstackv1beta1.NovaCell {
	labels := template.AppLabels(instance.Name, AppLabel)

	return &openstackv1beta1.NovaCell{
		ObjectMeta: metav1.ObjectMeta{
			Name:      template.Combine(instance.Name, spec.Name),
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: spec,
	}
}

func CellDBSyncJob(instance *openstackv1beta1.NovaCell, env []corev1.EnvVar, volumes []corev1.Volume, containerImage string) *batchv1.Job {
	labels := template.AppLabels(instance.Name, AppLabel)

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
					template.MustRenderFile(AppLabel, "cell-db-sync.sh", nil),
				},
				Env: env,
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "etc-nova",
						SubPath:   "nova.conf",
						MountPath: "/etc/nova/nova.conf",
					},
				},
			},
		},
		Volumes: volumes,
	})

	job.Name = template.Combine(instance.Name, "db-sync")

	return job
}

func EnsureCell(ctx context.Context, c client.Client, intended *openstackv1beta1.NovaCell, log logr.Logger) error {
	hash, err := template.ObjectHash(intended)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}

	found := &openstackv1beta1.NovaCell{}
	if err := c.Get(ctx, client.ObjectKeyFromObject(intended), found); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		template.SetAppliedHash(intended, hash)

		log.Info("Creating NovaCell", "Name", intended.Name)
		return c.Create(ctx, intended)
	} else if !template.MatchesAppliedHash(found, hash) {
		found.Spec = intended.Spec
		template.SetAppliedHash(found, hash)

		log.Info("Updating NovaCell", "Name", intended.Name)
		return c.Update(ctx, found)
	}

	return nil
}

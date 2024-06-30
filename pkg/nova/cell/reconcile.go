package cell

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
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

	spec := instance.Spec.DBSyncJob

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
				Resources:    spec.Resources,
				VolumeMounts: volumeMounts,
			},
		},
		NodeSelector: spec.NodeSelector,
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
	return template.Ensure(ctx, c, instance, log, func(intended *openstackv1beta1.NovaCell) {
		instance.Spec = intended.Spec
	})
}

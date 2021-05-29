package mariadb

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	ClusterComponentLabel = "cluster"
)

func ClusterStatefulSet(instance *openstackv1beta1.MariaDB, configHash string) *appsv1.StatefulSet {
	labels := template.Labels(instance.Name, AppLabel, ClusterComponentLabel)

	runAsUser := int64(1001)

	probe := &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &corev1.ExecAction{
				Command: []string{"bash", "-c", template.MustRenderFile(AppLabel, "probe.sh", nil)},
			},
		},
		FailureThreshold:    3,
		InitialDelaySeconds: 30,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		TimeoutSeconds:      1,
	}

	sts := template.GenericStatefulSet(template.Component{
		Namespace: instance.Namespace,
		Labels:    labels,
		SecurityContext: &corev1.PodSecurityContext{
			FSGroup: &runAsUser,
		},
		Containers: []corev1.Container{
			{
				Name:  "mariadb",
				Image: instance.Spec.Image,
				SecurityContext: &corev1.SecurityContext{
					RunAsUser: &runAsUser,
				},
				Env: []corev1.EnvVar{
					template.EnvVar("CONFIG_HASH", configHash),
					template.EnvVar("BITNAMI_DEBUG", "false"),
					template.SecretEnvVar("MARIADB_ROOT_PASSWORD", instance.Name, "password"),
				},
				Ports: []corev1.ContainerPort{
					{Name: "mysql", ContainerPort: 3306},
				},
				LivenessProbe:  probe,
				ReadinessProbe: probe,
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "config",
						MountPath: "/opt/bitnami/mariadb/conf/my.cnf",
						SubPath:   "my.cnf",
					},
					{
						Name:      "data",
						MountPath: "/bitnami/mariadb",
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			template.ConfigMapVolume("config", instance.Name, nil),
		},
		VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			template.PersistentVolumeClaim("data", labels, instance.Spec.Volume),
		},
	})

	return sts
}

func ClusterService(instance *openstackv1beta1.MariaDB) *corev1.Service {
	labels := template.Labels(instance.Name, AppLabel, ClusterComponentLabel)

	svc := template.GenericService(instance.Name, instance.Namespace, labels)
	svc.Spec.Ports = []corev1.ServicePort{
		{Name: "mysql", Port: 3306},
	}

	return svc
}

func ClusterHeadlessService(instance *openstackv1beta1.MariaDB) *corev1.Service {
	svc := ClusterService(instance)
	svc.Name = template.HeadlessServiceName(instance.Name)
	svc.Spec.ClusterIP = corev1.ClusterIPNone

	return svc
}

func EnsureCluster(ctx context.Context, c client.Client, intended *openstackv1beta1.MariaDB, log logr.Logger) error {
	hash, err := template.ObjectHash(intended)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}

	found := &openstackv1beta1.MariaDB{}
	if err := c.Get(ctx, client.ObjectKeyFromObject(intended), found); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		template.SetAppliedHash(intended, hash)

		log.Info("Creating MariaDB", "Name", intended.Name)
		return c.Create(ctx, intended)
	} else if !template.MatchesAppliedHash(found, hash) {
		found.Spec = intended.Spec
		template.SetAppliedHash(found, hash)

		log.Info("Updating MariaDB", "Name", intended.Name)
		return c.Update(ctx, found)
	}

	return nil
}

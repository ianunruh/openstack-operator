package mariadb

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	ClusterComponentLabel = "cluster"
)

func ClusterStatefulSet(instance *openstackv1beta1.MariaDB, configHash string) *appsv1.StatefulSet {
	labels := template.Labels(instance.Name, AppLabel, ClusterComponentLabel)

	runAsUser := int64(1001)

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
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "data",
					Labels: labels,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
					Resources:        template.StorageResources(instance.Spec.Volume.Capacity),
					StorageClassName: instance.Spec.Volume.StorageClass,
				},
			},
		},
	})

	return sts
}

func ClusterService(instance *openstackv1beta1.MariaDB) *corev1.Service {
	labels := template.Labels(instance.Name, AppLabel, ClusterComponentLabel)

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{Name: "mysql", Port: 3306},
			},
		},
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

		log.Info("Creating MariaDB cluster", "Name", intended.Name)
		return c.Create(ctx, intended)
	} else if !template.MatchesAppliedHash(found, hash) {
		found.Spec = intended.Spec
		template.SetAppliedHash(found, hash)

		log.Info("Updating MariaDB cluster", "Name", intended.Name)
		return c.Update(ctx, found)
	}

	return nil
}

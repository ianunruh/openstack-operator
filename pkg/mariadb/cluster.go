package mariadb

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/prometheus"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	ClusterComponentLabel = "cluster"
)

func ClusterStatefulSet(instance *openstackv1beta1.MariaDB, configHash string) *appsv1.StatefulSet {
	labels := template.Labels(instance.Name, AppLabel, ClusterComponentLabel)

	runAsUser := int64(1001)

	probe := &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			Exec: &corev1.ExecAction{
				Command: []string{"bash", "-c", template.MustReadFile(AppLabel, "probe.sh")},
			},
		},
		FailureThreshold:    3,
		InitialDelaySeconds: 30,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		TimeoutSeconds:      1,
	}

	containers := []corev1.Container{
		{
			Name:  "mariadb",
			Image: instance.Spec.Image,
			Env: []corev1.EnvVar{
				template.EnvVar("CONFIG_HASH", configHash),
				template.EnvVar("BITNAMI_DEBUG", "false"),
				template.SecretEnvVar("MARIADB_ROOT_PASSWORD", instance.Name, "password"),
			},
			Ports: []corev1.ContainerPort{
				{Name: "mysql", ContainerPort: 3306},
			},
			LivenessProbe: probe,
			StartupProbe:  probe,
			Resources:     instance.Spec.Resources,
			VolumeMounts: []corev1.VolumeMount{
				template.SubPathVolumeMount("config", "/opt/bitnami/mariadb/conf/my.cnf", "my.cnf"),
				template.VolumeMount("data", "/bitnami/mariadb"),
			},
		},
	}

	if promSpec := instance.Spec.Prometheus; promSpec != nil {
		containers = append(containers, exporterContainer(promSpec.Exporter, instance.Name))
	}

	sts := template.GenericStatefulSet(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		NodeSelector: instance.Spec.NodeSelector,
		SecurityContext: &corev1.PodSecurityContext{
			FSGroup:   &runAsUser,
			RunAsUser: &runAsUser,
		},
		Containers: containers,
		Volumes: []corev1.Volume{
			template.ConfigMapVolume("config", instance.Name, nil),
		},
		VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			template.PersistentVolumeClaim("data", labels, instance.Spec.Volume),
		},
	})

	return sts
}

func exporterContainer(spec openstackv1beta1.MariaDBExporterSpec, secret string) corev1.Container {
	return corev1.Container{
		Name:  "exporter",
		Image: spec.Image,
		Command: []string{
			"bash",
			"-c",
			template.MustReadFile(AppLabel, "start-exporter.sh"),
		},
		Env: []corev1.EnvVar{
			template.SecretEnvVar("MARIADB_ROOT_PASSWORD", secret, "password"),
		},
		Ports: []corev1.ContainerPort{
			{Name: "metrics", ContainerPort: 9104},
		},
		Resources: spec.Resources,
	}
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
	extraPorts := []corev1.ServicePort{
		{Name: "metrics", Port: 9104},
	}

	svc := ClusterService(instance)
	svc.Name = template.HeadlessServiceName(instance.Name)
	svc.Spec.ClusterIP = corev1.ClusterIPNone
	svc.Spec.Ports = append(svc.Spec.Ports, extraPorts...)

	return svc
}

func ClusterServiceMonitor(instance *openstackv1beta1.MariaDB) *unstructured.Unstructured {
	return prometheus.ServiceMonitor(prometheus.ServiceMonitorParams{
		Name:          instance.Name,
		Namespace:     instance.Namespace,
		NameLabel:     AppLabel,
		InstanceLabel: instance.Name,
	})
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

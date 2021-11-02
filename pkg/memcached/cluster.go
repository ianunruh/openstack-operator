package memcached

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
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	ClusterComponentLabel = "cluster"
)

func ClusterStatefulSet(instance *openstackv1beta1.Memcached) *appsv1.StatefulSet {
	labels := template.Labels(instance.Name, AppLabel, ClusterComponentLabel)

	runAsUser := int64(1001)

	containers := []corev1.Container{
		{
			Name:  "memcached",
			Image: instance.Spec.Image,
			Args: []string{
				"/run.sh",
				"-e/cache-state/memory_file",
			},
			Ports: []corev1.ContainerPort{
				{Name: "memcached", ContainerPort: 11211},
			},
			Resources: instance.Spec.Resources,
			VolumeMounts: []corev1.VolumeMount{
				template.VolumeMount("data", "/cache-state"),
				template.VolumeMount("tmp", "/tmp"),
			},
		},
	}

	if promSpec := instance.Spec.Prometheus; promSpec != nil {
		containers = append(containers, exporterContainer(promSpec.Exporter))
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
			template.EmptyDirVolume("tmp"),
		},
		VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			template.PersistentVolumeClaim("data", labels, instance.Spec.Volume),
		},
	})

	return sts
}

func exporterContainer(spec openstackv1beta1.MemcachedExporterSpec) corev1.Container {
	return corev1.Container{
		Name:            "exporter",
		Image:           spec.Image,
		ImagePullPolicy: corev1.PullAlways,
		Ports: []corev1.ContainerPort{
			{Name: "metrics", ContainerPort: 9150},
		},
		Resources: spec.Resources,
	}
}

func ClusterService(instance *openstackv1beta1.Memcached) *corev1.Service {
	labels := template.Labels(instance.Name, AppLabel, ClusterComponentLabel)

	svc := template.GenericService(instance.Name, instance.Namespace, labels)
	svc.Spec.Ports = []corev1.ServicePort{
		{Name: "memcached", Port: 11211},
	}

	return svc
}

func ClusterHeadlessService(instance *openstackv1beta1.Memcached) *corev1.Service {
	svc := ClusterService(instance)
	svc.Name = template.HeadlessServiceName(instance.Name)
	svc.Spec.ClusterIP = corev1.ClusterIPNone

	return svc
}

type serviceMonitorOptions struct {
	Name      string
	Namespace string
}

func ClusterServiceMonitor(instance *openstackv1beta1.Memcached) *unstructured.Unstructured {
	manifest := template.MustRenderFile(AppLabel, "servicemonitor.yaml", serviceMonitorOptions{
		Name:      instance.Name,
		Namespace: instance.Namespace,
	})

	res := template.MustDecodeManifest(manifest)
	res.SetNamespace(instance.Namespace)

	return res
}

func EnsureCluster(ctx context.Context, c client.Client, intended *openstackv1beta1.Memcached, log logr.Logger) error {
	hash, err := template.ObjectHash(intended)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}

	found := &openstackv1beta1.Memcached{}
	if err := c.Get(ctx, client.ObjectKeyFromObject(intended), found); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		template.SetAppliedHash(intended, hash)

		log.Info("Creating Memcached", "Name", intended.Name)
		return c.Create(ctx, intended)
	} else if !template.MatchesAppliedHash(found, hash) {
		found.Spec = intended.Spec
		template.SetAppliedHash(found, hash)

		log.Info("Updating Memcached", "Name", intended.Name)
		return c.Update(ctx, found)
	}

	return nil
}

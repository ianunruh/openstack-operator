package glance

import (
	"slices"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/rookceph"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	APIComponentLabel = "api"
)

func APIDeployment(instance *openstackv1beta1.Glance, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.Deployment {
	labels := template.Labels(instance.Name, AppLabel, APIComponentLabel)

	probe := &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.FromInt(9292),
			},
		},
		InitialDelaySeconds: 5,
		PeriodSeconds:       10,
		TimeoutSeconds:      5,
	}

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-glance", "/etc/glance/glance-api.conf", "glance-api.conf"),
		template.SubPathVolumeMount("etc-glance", "/var/lib/kolla/config_files/config.json", "kolla.json"),
	}

	var deployStrategyType appsv1.DeploymentStrategyType

	cephSecrets := rookceph.NewClientSecretAppender(&volumes, &volumeMounts)
	for _, backend := range instance.Spec.Backends {
		if cephSpec := backend.Ceph; cephSpec != nil {
			cephSecrets.Append(cephSpec.Secret)
		} else if pvcSpec := backend.PVC; pvcSpec != nil {
			pvcName := template.Combine(instance.Name, backend.Name)

			volumeMounts = append(volumeMounts, corev1.VolumeMount{
				Name:      "images",
				MountPath: imageBackendPath(backend.Name),
			})
			volumes = append(volumes, template.PersistentVolume("images", pvcName))

			if isReadWriteOnce(pvcSpec.AccessModes) {
				// avoid state where new pods cannot mount PVC due to exclusivity
				deployStrategyType = appsv1.RecreateDeploymentStrategyType
			}
		}
	}

	deploy := template.GenericDeployment(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		Replicas:     instance.Spec.API.Replicas,
		NodeSelector: instance.Spec.API.NodeSelector,
		Affinity: &corev1.Affinity{
			PodAntiAffinity: template.NodePodAntiAffinity(labels),
		},
		Containers: []corev1.Container{
			{
				Name:    "api",
				Image:   instance.Spec.Image,
				Command: []string{"/usr/local/bin/kolla_start"},
				Env:     env,
				Ports: []corev1.ContainerPort{
					{Name: "http", ContainerPort: 9292},
				},
				LivenessProbe: probe,
				StartupProbe:  probe,
				Resources:     instance.Spec.API.Resources,
				VolumeMounts:  volumeMounts,
			},
		},
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser: &appUID,
			FSGroup:   &appUID,
		},
		Volumes: volumes,
	})

	deploy.Name = template.Combine(instance.Name, "api")
	deploy.Spec.Strategy.Type = deployStrategyType

	return deploy
}

func APIService(instance *openstackv1beta1.Glance) *corev1.Service {
	labels := template.Labels(instance.Name, AppLabel, APIComponentLabel)
	name := template.Combine(instance.Name, "api")

	svc := template.GenericService(name, instance.Namespace, labels)
	svc.Spec.Ports = []corev1.ServicePort{
		{Name: "http", Port: 9292},
	}

	return svc
}

func APIIngress(instance *openstackv1beta1.Glance) *netv1.Ingress {
	labels := template.Labels(instance.Name, AppLabel, APIComponentLabel)
	name := template.Combine(instance.Name, APIComponentLabel)

	ingress := template.GenericIngress(name, instance.Namespace, instance.Spec.API.Ingress, labels)
	ingress.Annotations = template.MergeStringMaps(ingress.Annotations, map[string]string{
		"nginx.ingress.kubernetes.io/proxy-body-size": "0",
	})

	return ingress
}

func isReadWriteOnce(accessModes []corev1.PersistentVolumeAccessMode) bool {
	return slices.Contains(accessModes, corev1.ReadWriteOnce) || slices.Contains(accessModes, corev1.ReadWriteOncePod)
}

package glance

import (
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

func APIDeployment(instance *openstackv1beta1.Glance, configHash string) *appsv1.Deployment {
	labels := template.Labels(instance.Name, AppLabel, APIComponentLabel)

	keystoneSecret := template.Combine(instance.Name, "keystone")

	probe := &corev1.Probe{
		Handler: corev1.Handler{
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
	}

	volumes := []corev1.Volume{
		template.ConfigMapVolume("etc-glance", instance.Name, nil),
	}

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
		}
	}

	deploy := template.GenericDeployment(template.Component{
		Namespace: instance.Namespace,
		Labels:    labels,
		Replicas:  instance.Spec.API.Replicas,
		Containers: []corev1.Container{
			{
				Name:  "api",
				Image: instance.Spec.Image,
				Command: []string{
					"glance-api",
					"--config-file=/etc/glance/glance-api.conf",
				},
				Env: []corev1.EnvVar{
					template.EnvVar("CONFIG_HASH", configHash),
					template.SecretEnvVar("OS_DATABASE__CONNECTION", instance.Spec.Database.Secret, "connection"),
					template.SecretEnvVar("OS_KEYSTONE_AUTHTOKEN__PASSWORD", keystoneSecret, "OS_PASSWORD"),
				},
				Ports: []corev1.ContainerPort{
					{Name: "http", ContainerPort: 9292},
				},
				LivenessProbe: probe,
				StartupProbe:  probe,
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

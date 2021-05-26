package glance

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	}

	runAsUser := int64(64062)

	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "etc-glance",
			SubPath:   "glance-api.conf",
			MountPath: "/etc/glance/glance-api.conf",
		},
	}

	volumes := []corev1.Volume{
		template.ConfigMapVolume("etc-glance", instance.Name, nil),
	}

	if cephSpec := instance.Spec.Storage.RookCeph; cephSpec != nil {
		volumeMounts = append(volumeMounts, rookceph.ClientVolumeMounts("etc-ceph")...)
		volumes = append(volumes, template.SecretVolume("etc-ceph", cephSpec.Secret, nil))
	} else if instance.Spec.Storage.Volume != nil {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "images",
			MountPath: "/var/lib/glance/images",
		})
		volumes = append(volumes, template.PersistentVolume("images", instance.Name))
	}

	deploy := template.GenericDeployment(template.Component{
		Namespace: instance.Namespace,
		Labels:    labels,
		Replicas:  instance.Spec.API.Replicas,
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser: &runAsUser,
			FSGroup:   &runAsUser,
		},
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
				LivenessProbe:  probe,
				ReadinessProbe: probe,
				VolumeMounts:   volumeMounts,
			},
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

	spec := instance.Spec.API.Ingress

	prefixPathType := netv1.PathTypePrefix

	svcName := template.Combine(instance.Name, "api")

	annotations := map[string]string{
		"nginx.ingress.kubernetes.io/proxy-body-size": "0",
	}

	ingress := &netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        template.Combine(instance.Name, "api"),
			Namespace:   instance.Namespace,
			Labels:      labels,
			Annotations: template.MergeStringMaps(annotations, spec.Annotations),
		},
		Spec: netv1.IngressSpec{
			TLS: []netv1.IngressTLS{
				{
					SecretName: template.Combine(instance.Name, "api-ingress-tls"),
					Hosts:      []string{spec.Host},
				},
			},
			Rules: []netv1.IngressRule{
				{
					Host: spec.Host,
					IngressRuleValue: netv1.IngressRuleValue{
						HTTP: &netv1.HTTPIngressRuleValue{
							Paths: []netv1.HTTPIngressPath{
								{
									PathType: &prefixPathType,
									Path:     "/",
									Backend:  template.IngressServiceBackend(svcName, "http"),
								},
							},
						},
					},
				},
			},
		},
	}

	return ingress
}

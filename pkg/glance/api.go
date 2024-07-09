package glance

import (
	"fmt"
	"slices"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/pki"
	"github.com/ianunruh/openstack-operator/pkg/rookceph"
	"github.com/ianunruh/openstack-operator/pkg/template"
	"github.com/ianunruh/openstack-operator/pkg/tlsproxy"
)

const (
	APIComponentLabel = "api"
)

func APIDeployment(instance *openstackv1beta1.Glance, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.Deployment {
	labels := template.Labels(instance.Name, AppLabel, APIComponentLabel)

	spec := instance.Spec.API

	probe := &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path:   "/healthcheck",
				Port:   intstr.FromInt(9292),
				Scheme: pki.HTTPActionScheme(spec.TLS),
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

	pki.AppendTLSClientVolumes(instance.Spec.TLS, &volumes, &volumeMounts)

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

	apiContainer := corev1.Container{
		Name:         "api",
		Image:        spec.Image,
		Command:      []string{"/usr/local/bin/kolla_start"},
		Env:          env,
		Resources:    spec.Resources,
		VolumeMounts: volumeMounts,
	}

	var containers []corev1.Container

	if spec.TLS.Secret == "" {
		apiContainer.Ports = []corev1.ContainerPort{
			{Name: "http", ContainerPort: 9292},
		}
		apiContainer.LivenessProbe = probe
		apiContainer.StartupProbe = probe
	} else {
		tlsProxyVolumeMounts := tlsproxy.VolumeMounts("etc-glance", "tlsproxy.conf")
		tlsproxy.AppendTLSServerVolumes(spec.TLS, &volumes, &tlsProxyVolumeMounts)

		containers = append(containers,
			tlsproxy.Container(9292, spec.TLSProxy, probe, tlsProxyVolumeMounts))
	}

	containers = append(containers, apiContainer)

	deploy := template.GenericDeployment(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		Replicas:     spec.Replicas,
		NodeSelector: spec.NodeSelector,
		Affinity: &corev1.Affinity{
			PodAntiAffinity: template.NodePodAntiAffinity(labels),
		},
		Containers: containers,
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

	spec := instance.Spec.API

	ingress := template.GenericIngressWithTLS(name, instance.Namespace, spec.Ingress, spec.TLS, labels)
	ingress.Annotations = template.MergeStringMaps(ingress.Annotations, map[string]string{
		"nginx.ingress.kubernetes.io/proxy-body-size": "0",
	})

	return ingress
}

func APIInternalURL(instance *openstackv1beta1.Glance) string {
	scheme := "http"
	if instance.Spec.API.TLS.Secret != "" {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s-api.%s.svc:9292", scheme, instance.Name, instance.Namespace)
}

func APIPublicURL(instance *openstackv1beta1.Glance) string {
	if instance.Spec.API.Ingress == nil {
		return APIInternalURL(instance)
	}
	return fmt.Sprintf("https://%s", instance.Spec.API.Ingress.Host)
}

func isReadWriteOnce(accessModes []corev1.PersistentVolumeAccessMode) bool {
	return slices.Contains(accessModes, corev1.ReadWriteOnce) || slices.Contains(accessModes, corev1.ReadWriteOncePod)
}

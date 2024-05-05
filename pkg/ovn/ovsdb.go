package ovn

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	OVSDBNorth = "ovsdb-nb"
	OVSDBSouth = "ovsdb-sb"

	OVSDBNorthPort = 6641
	OVSDBSouthPort = 6642
)

func OVSDBStatefulSet(instance *openstackv1beta1.OVNControlPlane, component string, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.StatefulSet {
	labels := template.Labels(instance.Name, AppLabel, component)

	port := ovsdbPort(component)
	spec := ovsdbSpec(instance, component)

	kollaConfig := fmt.Sprintf("kolla-%s.json", component)

	volumeMounts := []corev1.VolumeMount{
		template.VolumeMount("data", "/var/lib/ovn"),
		template.SubPathVolumeMount("etc-ovn", "/var/lib/kolla/config_files/config.json", kollaConfig),
	}

	probe := &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.FromInt(int(port)),
			},
		},
		InitialDelaySeconds: 5,
		PeriodSeconds:       10,
		TimeoutSeconds:      3,
	}

	ds := template.GenericStatefulSet(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		NodeSelector: spec.NodeSelector,
		Containers: []corev1.Container{
			{
				Name:    "ovsdb",
				Image:   spec.Image,
				Command: []string{"/usr/local/bin/kolla_start"},
				Env:     env,
				Ports: []corev1.ContainerPort{
					{Name: "ovsdb", ContainerPort: port},
				},
				LivenessProbe: probe,
				StartupProbe:  probe,
				Resources:     spec.Resources,
				VolumeMounts:  volumeMounts,
			},
		},
		VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			template.PersistentVolumeClaim("data", labels, spec.Volume),
		},
		Volumes: volumes,
	})

	ds.Name = template.Combine(instance.Name, component)

	return ds
}

func OVSDBService(instance *openstackv1beta1.OVNControlPlane, component string) *corev1.Service {
	labels := template.Labels(instance.Name, AppLabel, component)
	name := template.Combine(instance.Name, component)

	port := ovsdbPort(component)

	svc := template.GenericService(name, instance.Namespace, labels)
	svc.Spec.Ports = []corev1.ServicePort{
		{Name: "ovsdb", Port: port},
	}

	return svc
}

func ovsdbPort(component string) int32 {
	if component == OVSDBNorth {
		return OVSDBNorthPort
	}
	return OVSDBSouthPort
}

func ovsdbSpec(instance *openstackv1beta1.OVNControlPlane, component string) openstackv1beta1.OVNDBSpec {
	if component == OVSDBNorth {
		return instance.Spec.OVSDBNorth
	}
	return instance.Spec.OVSDBSouth
}

func OVSDBConnectionConfigMap(instance *openstackv1beta1.OVNControlPlane, northSvc, southSvc *corev1.Service) *corev1.ConfigMap {
	name := template.Combine(instance.Name, "ovsdb")
	labels := template.AppLabels(instance.Name, AppLabel)

	cm := template.GenericConfigMap(name, instance.Namespace, labels)
	cm.Data["OVN_NB_CONNECTION"] = fmt.Sprintf("tcp:%s:6641", northSvc.Spec.ClusterIP)
	cm.Data["OVN_SB_CONNECTION"] = fmt.Sprintf("tcp:%s:6642", southSvc.Spec.ClusterIP)

	return cm
}

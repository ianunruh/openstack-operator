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

func OVSDBStatefulSet(instance *openstackv1beta1.OVNControlPlane, component string) *appsv1.StatefulSet {
	labels := template.Labels(instance.Name, AppLabel, component)

	port := ovsdbPort(component)
	spec := ovsdbSpec(instance, component)
	startScript := ovsdbStartScript(component)

	probe := &corev1.Probe{
		Handler: corev1.Handler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.FromInt(int(port)),
			},
		},
		InitialDelaySeconds: 5,
		PeriodSeconds:       10,
		TimeoutSeconds:      3,
	}

	ds := template.GenericStatefulSet(template.Component{
		Namespace: instance.Namespace,
		Labels:    labels,
		Containers: []corev1.Container{
			{
				Name:  "ovsdb",
				Image: instance.Spec.Image,
				Command: []string{
					"bash",
					"-c",
					startScript,
				},
				Ports: []corev1.ContainerPort{
					{Name: "ovsdb", ContainerPort: port},
				},
				LivenessProbe: probe,
				StartupProbe:  probe,
				VolumeMounts: []corev1.VolumeMount{
					template.VolumeMount("data", "/var/lib/ovn"),
				},
			},
		},
		VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			template.PersistentVolumeClaim("data", labels, spec.Volume),
		},
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

func ovsdbStartScript(component string) string {
	if component == OVSDBNorth {
		return template.MustReadFile(AppLabel, "start-ovsdb-nb.sh")
	}
	return template.MustReadFile(AppLabel, "start-ovsdb-sb.sh")
}

func ovsdbSpec(instance *openstackv1beta1.OVNControlPlane, component string) *openstackv1beta1.OVSDBSpec {
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

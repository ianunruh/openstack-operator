package nova

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	LibvirtdComponentLabel = "libvirtd"
)

func LibvirtdConfigMap(instance *openstackv1beta1.Nova) *corev1.ConfigMap {
	labels := template.Labels(instance.Name, AppLabel, LibvirtdComponentLabel)
	name := template.Combine(instance.Name, "libvirtd")
	cm := template.GenericConfigMap(name, instance.Namespace, labels)

	cm.Data["libvirtd.conf"] = template.MustRenderFile(AppLabel, "libvirtd.conf", nil)
	cm.Data["qemu.conf"] = template.MustRenderFile(AppLabel, "qemu.conf", nil)

	return cm
}

func LibvirtdDaemonSet(instance *openstackv1beta1.Nova, configHash string) *appsv1.DaemonSet {
	labels := template.Labels(instance.Name, AppLabel, LibvirtdComponentLabel)

	configMapName := template.Combine(instance.Name, "libvirtd")

	runAsRootUser := int64(0)
	privileged := true

	mountPropagation := corev1.MountPropagationBidirectional

	probe := &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &corev1.ExecAction{
				Command: []string{"bash", "-c", "/usr/bin/virsh list"},
			},
		},
	}

	ds := template.GenericDaemonSet(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		NodeSelector: instance.Spec.Compute.NodeSelector,
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser: &runAsRootUser,
		},
		Containers: []corev1.Container{
			{
				Name:  "libvirtd",
				Image: instance.Spec.Libvirtd.Image,
				Command: []string{
					"bash",
					"-c",
					template.MustRenderFile(AppLabel, "libvirtd-start.sh", nil),
				},
				Lifecycle: &corev1.Lifecycle{
					PreStop: &corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{"bash", "-c", "kill $(cat /var/run/libvirtd.pid)"},
						},
					},
				},
				Env: []corev1.EnvVar{
					template.EnvVar("CONFIG_HASH", configHash),
				},
				ReadinessProbe: probe,
				LivenessProbe:  probe,
				SecurityContext: &corev1.SecurityContext{
					Privileged: &privileged,
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "etc-libvirt",
						MountPath: "/etc/libvirt/libvirtd.conf",
						SubPath:   "libvirtd.conf",
					},
					{
						Name:      "etc-libvirt",
						MountPath: "/etc/libvirt/qemu.conf",
						SubPath:   "qemu.conf",
					},
					{
						Name:      "pod-tmp",
						MountPath: "/tmp",
					},
					{
						Name:      "host-dev",
						MountPath: "/dev",
					},
					{
						Name:      "host-etc-libvirt-qemu",
						MountPath: "/etc/libvirt/qemu",
					},
					{
						Name:      "host-etc-machine-id",
						MountPath: "/etc/machine-id",
						ReadOnly:  true,
					},
					{
						Name:      "host-lib-modules",
						MountPath: "/lib/modules",
						ReadOnly:  true,
					},
					{
						Name:      "host-run",
						MountPath: "/run",
					},
					{
						Name:      "host-sys-fs-cgroup",
						MountPath: "/sys/fs/cgroup",
					},
					{
						Name:             "host-var-lib-libvirt",
						MountPath:        "/var/lib/libvirt",
						MountPropagation: &mountPropagation,
					},
					{
						Name:             "host-var-lib-nova",
						MountPath:        "/var/lib/nova",
						MountPropagation: &mountPropagation,
					},
					{
						Name:      "host-var-log-libvirt",
						MountPath: "/var/log/libvirt",
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			template.ConfigMapVolume("etc-libvirt", configMapName, nil),
			template.EmptyDirVolume("pod-tmp"),
			template.HostPathVolume("host-dev", "/dev"),
			template.HostPathVolume("host-etc-libvirt-qemu", "/etc/libvirt/qemu"),
			template.HostPathVolume("host-etc-machine-id", "/etc/machine-id"),
			template.HostPathVolume("host-lib-modules", "/lib/modules"),
			template.HostPathVolume("host-run", "/run"),
			template.HostPathVolume("host-sys-fs-cgroup", "/sys/fs/cgroup"),
			template.HostPathVolume("host-var-lib-libvirt", "/var/lib/libvirt"),
			template.HostPathVolume("host-var-lib-nova", "/var/lib/nova"),
			template.HostPathVolume("host-var-log-libvirt", "/var/log/libvirt"),
		},
	})

	ds.Name = template.Combine(instance.Name, "libvirtd")

	ds.Spec.Template.Spec.DNSPolicy = corev1.DNSClusterFirstWithHostNet
	ds.Spec.Template.Spec.HostIPC = true
	ds.Spec.Template.Spec.HostNetwork = true
	ds.Spec.Template.Spec.HostPID = true

	return ds
}

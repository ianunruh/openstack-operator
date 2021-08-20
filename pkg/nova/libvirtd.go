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

	cm.Data["libvirtd.conf"] = template.MustReadFile(AppLabel, "libvirtd.conf")
	cm.Data["qemu.conf"] = template.MustReadFile(AppLabel, "qemu.conf")

	return cm
}

func LibvirtdDaemonSet(instance *openstackv1beta1.Nova, env []corev1.EnvVar, volumeMounts []corev1.VolumeMount, volumes []corev1.Volume) *appsv1.DaemonSet {
	labels := template.Labels(instance.Name, AppLabel, LibvirtdComponentLabel)

	configMapName := template.Combine(instance.Name, "libvirtd")

	runAsRootUser := int64(0)
	privileged := true

	probe := &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &corev1.ExecAction{
				Command: []string{"bash", "-c", "/usr/bin/virsh list"},
			},
		},
		InitialDelaySeconds: 5,
		PeriodSeconds:       10,
		TimeoutSeconds:      5,
	}

	extraVolumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-libvirt", "/etc/libvirt/libvirtd.conf", "libvirtd.conf"),
		template.SubPathVolumeMount("etc-libvirt", "/etc/libvirt/qemu.conf", "qemu.conf"),
		template.VolumeMount("pod-tmp", "/tmp"),
		template.VolumeMount("host-dev", "/dev"),
		template.VolumeMount("host-etc-libvirt-qemu", "/etc/libvirt/qemu"),
		template.ReadOnlyVolumeMount("host-etc-machine-id", "/etc/machine-id"),
		template.ReadOnlyVolumeMount("host-lib-modules", "/lib/modules"),
		template.VolumeMount("host-run", "/run"),
		template.VolumeMount("host-sys-fs-cgroup", "/sys/fs/cgroup"),
		template.BidirectionalVolumeMount("host-var-lib-libvirt", "/var/lib/libvirt"),
		template.BidirectionalVolumeMount("host-var-lib-nova", "/var/lib/nova"),
		template.VolumeMount("host-var-log-libvirt", "/var/log/libvirt"),
	}

	extraVolumes := []corev1.Volume{
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
					template.MustReadFile(AppLabel, "libvirtd-start.sh"),
				},
				Lifecycle: &corev1.Lifecycle{
					PreStop: &corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{
								"bash",
								"-c",
								"kill $(cat /var/run/libvirtd.pid)",
							},
						},
					},
				},
				Env:           env,
				LivenessProbe: probe,
				StartupProbe:  probe,
				SecurityContext: &corev1.SecurityContext{
					Privileged: &privileged,
				},
				VolumeMounts: append(volumeMounts, extraVolumeMounts...),
			},
		},
		Volumes: append(volumes, extraVolumes...),
	})

	ds.Name = template.Combine(instance.Name, "libvirtd")

	ds.Spec.Template.Spec.DNSPolicy = corev1.DNSClusterFirstWithHostNet
	ds.Spec.Template.Spec.HostIPC = true
	ds.Spec.Template.Spec.HostNetwork = true
	ds.Spec.Template.Spec.HostPID = true

	return ds
}

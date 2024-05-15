package nova

import (
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	ComputeSSHComponentLabel = "compute-ssh"

	ComputeSSHPort = 2022
)

func ComputeSSHDaemonSet(instance *openstackv1beta1.NovaComputeSet, env []corev1.EnvVar, volumes []corev1.Volume) *appsv1.DaemonSet {
	labels := template.Labels(instance.Name, AppLabel, ComputeSSHComponentLabel)

	spec := instance.Spec.SSH

	defaultMode := int32(0400)

	env = append(env,
		template.EnvVar("NOVA_USER_UID", strconv.Itoa(int(appUID))),
		template.EnvVar("SSH_PORT", strconv.Itoa(ComputeSSHPort)))

	volumes = append(volumes,
		template.SecretVolume("ssh-keys", "nova-compute-ssh", &defaultMode),
		template.HostPathVolume("host-var-lib-nova", "/var/lib/nova"))

	volumeMounts := []corev1.VolumeMount{
		template.SubPathVolumeMount("etc-nova", "/scripts/compute-ssh.sh", "compute-ssh.sh"),
		template.SubPathVolumeMount("etc-nova", "/var/lib/kolla/config_files/config.json", "kolla-nova-compute-ssh.json"),
		template.VolumeMount("ssh-keys", "/tmp/ssh-keys"),
		template.BidirectionalVolumeMount("host-var-lib-nova", "/var/lib/nova"),
	}

	runAsRootUser := int64(0)
	privileged := true

	ds := template.GenericDaemonSet(template.Component{
		Namespace:    instance.Namespace,
		Labels:       labels,
		NodeSelector: instance.Spec.NodeSelector,
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser: &appUID,
		},
		Containers: []corev1.Container{
			{
				Name:    "ssh",
				Image:   spec.Image,
				Command: []string{"/usr/local/bin/kolla_start"},
				Env:     env,
				Ports: []corev1.ContainerPort{
					{Name: "ssh", ContainerPort: ComputeSSHPort},
				},
				Resources: spec.Resources,
				SecurityContext: &corev1.SecurityContext{
					RunAsUser:  &runAsRootUser,
					Privileged: &privileged,
				},
				VolumeMounts: volumeMounts,
			},
		},
		Volumes: volumes,
	})

	ds.Name = template.Combine(instance.Name, "compute-ssh")

	ds.Spec.Template.Spec.DNSPolicy = corev1.DNSClusterFirstWithHostNet
	ds.Spec.Template.Spec.HostNetwork = true

	return ds
}

func ComputeSSHKeypairSecret(instance *openstackv1beta1.Nova) (*corev1.Secret, error) {
	labels := template.Labels(instance.Name, AppLabel, ComputeSSHComponentLabel)
	name := template.Combine(instance.Name, "compute-ssh")

	return template.SSHKeypairSecret(name, instance.Namespace, labels)
}

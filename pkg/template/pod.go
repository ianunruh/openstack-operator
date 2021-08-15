package template

import (
	corev1 "k8s.io/api/core/v1"
)

func EnvVar(name, value string) corev1.EnvVar {
	return corev1.EnvVar{
		Name:  name,
		Value: value,
	}
}

func FieldEnvVar(name, fieldPath string) corev1.EnvVar {
	return corev1.EnvVar{
		Name: name,
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: fieldPath,
			},
		},
	}
}

func ConfigMapEnvVar(name, cmName, cmKey string) corev1.EnvVar {
	return corev1.EnvVar{
		Name: name,
		ValueFrom: &corev1.EnvVarSource{
			ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cmName,
				},
				Key: cmKey,
			},
		},
	}
}

func SecretEnvVar(name, secretName, secretKey string) corev1.EnvVar {
	return corev1.EnvVar{
		Name: name,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: secretName,
				},
				Key: secretKey,
			},
		},
	}
}

func EnvFromConfigMap(name string) corev1.EnvFromSource {
	return corev1.EnvFromSource{
		ConfigMapRef: &corev1.ConfigMapEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: name,
			},
		},
	}
}

func EnvFromSecret(name string) corev1.EnvFromSource {
	return EnvFromSecretPrefixed(name, "")
}

func EnvFromSecretPrefixed(name, prefix string) corev1.EnvFromSource {
	return corev1.EnvFromSource{
		Prefix: prefix,
		SecretRef: &corev1.SecretEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: name,
			},
		},
	}
}

func ConfigMapVolume(name, configMapName string, defaultMode *int32) corev1.Volume {
	return corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: configMapName,
				},
				DefaultMode: defaultMode,
			},
		},
	}
}

func SecretVolume(name, secretName string, defaultMode *int32) corev1.Volume {
	return corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName:  secretName,
				DefaultMode: defaultMode,
			},
		},
	}
}

func EmptyDirVolume(name string) corev1.Volume {
	return corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
}

func HostPathVolume(name, path string) corev1.Volume {
	return corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: path,
			},
		},
	}
}

func PersistentVolume(name, claimName string) corev1.Volume {
	return corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: claimName,
			},
		},
	}
}

func VolumeMount(name, mountPath string) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      name,
		MountPath: mountPath,
	}
}

func SubPathVolumeMount(name, mountPath, subPath string) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      name,
		MountPath: mountPath,
		SubPath:   subPath,
	}
}

func BidirectionalVolumeMount(name, mountPath string) corev1.VolumeMount {
	bidirectional := corev1.MountPropagationBidirectional

	return corev1.VolumeMount{
		Name:             name,
		MountPath:        mountPath,
		MountPropagation: &bidirectional,
	}
}

func ReadOnlyVolumeMount(name, mountPath string) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      name,
		MountPath: mountPath,
		ReadOnly:  true,
	}
}

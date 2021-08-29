package template

import (
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Component struct {
	Namespace            string
	Labels               map[string]string
	Replicas             int32
	Affinity             *corev1.Affinity
	NodeSelector         map[string]string
	InitContainers       []corev1.Container
	Containers           []corev1.Container
	SecurityContext      *corev1.PodSecurityContext
	Volumes              []corev1.Volume
	VolumeClaimTemplates []corev1.PersistentVolumeClaim
}

func GenericDaemonSet(component Component) *appsv1.DaemonSet {
	labels := component.Labels

	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels[InstanceLabel],
			Namespace: component.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Affinity:        component.Affinity,
					Containers:      component.Containers,
					InitContainers:  component.InitContainers,
					NodeSelector:    component.NodeSelector,
					SecurityContext: component.SecurityContext,
					Volumes:         component.Volumes,
				},
			},
		},
	}
}

func GenericDeployment(component Component) *appsv1.Deployment {
	labels := component.Labels

	replicas := component.Replicas
	if replicas == 0 {
		replicas = 1
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels[InstanceLabel],
			Namespace: component.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Affinity:        component.Affinity,
					Containers:      component.Containers,
					InitContainers:  component.InitContainers,
					NodeSelector:    component.NodeSelector,
					SecurityContext: component.SecurityContext,
					Volumes:         component.Volumes,
				},
			},
		},
	}
}

func GenericJob(component Component) *batchv1.Job {
	labels := component.Labels

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels[InstanceLabel],
			Namespace: component.Namespace,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Affinity:        component.Affinity,
					Containers:      component.Containers,
					InitContainers:  component.InitContainers,
					NodeSelector:    component.NodeSelector,
					SecurityContext: component.SecurityContext,
					Volumes:         component.Volumes,
					RestartPolicy:   corev1.RestartPolicyOnFailure,
				},
			},
		},
	}
}

func GenericStatefulSet(component Component) *appsv1.StatefulSet {
	labels := component.Labels

	replicas := component.Replicas
	if replicas == 0 {
		replicas = 1
	}

	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels[InstanceLabel],
			Namespace: component.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:    &replicas,
			ServiceName: HeadlessServiceName(labels[InstanceLabel]),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Affinity:        component.Affinity,
					Containers:      component.Containers,
					InitContainers:  component.InitContainers,
					NodeSelector:    component.NodeSelector,
					SecurityContext: component.SecurityContext,
					Volumes:         component.Volumes,
				},
			},
			VolumeClaimTemplates: component.VolumeClaimTemplates,
		},
	}
}

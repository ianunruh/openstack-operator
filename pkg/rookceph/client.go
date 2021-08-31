package rookceph

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/ianunruh/openstack-operator/pkg/template"
)

type ClientOptions struct {
	Name string
	Caps map[string]string
}

func Client(namespace string, opts ClientOptions) *unstructured.Unstructured {
	manifest := template.MustRenderFile(AppLabel, "ceph-client.yaml", opts)

	res := template.MustDecodeManifest(manifest)
	res.SetNamespace(namespace)

	return res
}

type clientSecretOptions struct {
	ClientName string
	MonHost    string
	SecretName string
}

type clientSecretKeyringOptions struct {
	ClientName string
	Keyring    string
}

func ClientSecret(name, namespace, clientName string, keyring []byte, monHosts []string) *corev1.Secret {
	// TODO labels
	secret := template.GenericSecret(name, namespace, nil)

	sort.Strings(monHosts)

	cephConf := template.MustRenderFile(AppLabel, "ceph.conf", clientSecretOptions{
		ClientName: clientName,
		MonHost:    strings.Join(monHosts, ","),
		SecretName: name,
	})

	keyringFile := template.MustRenderFile(AppLabel, "keyring", clientSecretKeyringOptions{
		ClientName: clientName,
		Keyring:    string(keyring),
	})

	secret.Data = map[string][]byte{
		"ceph.conf": []byte(cephConf),
		"keyring":   []byte(keyringFile),
	}

	return secret
}

func GetCephMonitorAddrs(ctx context.Context, c client.Client, namespace string) ([]string, error) {
	monLabels := labels.Set{"ceph_daemon_type": "mon"}

	listOpts := &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: monLabels.AsSelector(),
	}

	var monServices corev1.ServiceList
	if err := c.List(ctx, &monServices, listOpts); err != nil {
		return nil, err
	}

	var addrs []string

	if len(monServices.Items) > 0 {
		for _, svc := range monServices.Items {
			addrs = append(addrs, svc.Spec.ClusterIP)
		}

		return addrs, nil
	}

	// in rook-ceph clusters using host networking, no mon services are created.
	// fallback to mon pods
	var monPods corev1.PodList
	if err := c.List(ctx, &monPods, listOpts); err != nil {
		return nil, err
	}

	for _, pod := range monPods.Items {
		addrs = append(addrs, pod.Status.PodIP)
	}

	return addrs, nil
}

func GetCephClientSecret(ctx context.Context, c client.Client, name, namespace string) ([]byte, error) {
	keySecretName := template.Combine("rook-ceph-client", name)

	keySecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      keySecretName,
			Namespace: namespace,
		},
	}
	if err := c.Get(ctx, client.ObjectKeyFromObject(keySecret), keySecret); err != nil {
		return nil, err
	}

	keyring, ok := keySecret.Data[name]
	if !ok {
		return nil, fmt.Errorf("expected ceph client secret %s to have key %s", keySecretName, name)
	}

	return keyring, nil
}

func NewClientSecretAppender(volumes *[]corev1.Volume, volumeMounts *[]corev1.VolumeMount) *ClientSecretAppender {
	return &ClientSecretAppender{
		seenSecrets: make(map[string]bool),

		volumes:      volumes,
		volumeMounts: volumeMounts,
	}
}

type ClientSecretAppender struct {
	seenSecrets map[string]bool

	volumes      *[]corev1.Volume
	volumeMounts *[]corev1.VolumeMount
}

func (c *ClientSecretAppender) Append(name string) {
	mountPath := filepath.Join("/etc/ceph", name)

	*c.volumeMounts = append(*c.volumeMounts, ClientVolumeMounts(name, mountPath)...)
	*c.volumes = append(*c.volumes, template.SecretVolume(name, name, nil))
}

func ClientVolumeMounts(name, path string) []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      name,
			SubPath:   "ceph.conf",
			MountPath: filepath.Join(path, "ceph.conf"),
		},
		{
			Name:      name,
			SubPath:   "keyring",
			MountPath: filepath.Join(path, "keyring"),
		},
	}
}

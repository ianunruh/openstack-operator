package rookceph

import (
	"context"
	"fmt"
	"path/filepath"
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

	secret.StringData = map[string]string{
		"ceph.conf": template.MustRenderFile(AppLabel, "ceph.conf", clientSecretOptions{
			ClientName: clientName,
			MonHost:    strings.Join(monHosts, ","),
			SecretName: name,
		}),
		"keyring": template.MustRenderFile(AppLabel, "keyring", clientSecretKeyringOptions{
			ClientName: clientName,
			Keyring:    string(keyring),
		}),
	}

	return secret
}

func GetCephMonitorAddrs(ctx context.Context, c client.Client, namespace string) ([]string, error) {
	monLabels := labels.Set{"ceph_daemon_type": "mon"}

	// TODO prefer mon svc IPs if they exist
	var monPods corev1.PodList
	if err := c.List(ctx, &monPods, &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: monLabels.AsSelector(),
	}); err != nil {
		return nil, err
	}

	var addrs []string
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

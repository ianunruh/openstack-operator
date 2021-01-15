package rookceph

import (
	"context"
	"fmt"
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
	MonHost    string
	ClientName string
}

type clientSecretKeyringOptions struct {
	ClientName string
	Keyring    string
}

func ClientSecret(c client.Client, namespace, name, rookNamespace, clientName string) (*corev1.Secret, error) {
	keySecretName := template.Combine(clientName, "client-key")
	keySecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      keySecretName,
			Namespace: rookNamespace,
		},
	}
	if err := c.Get(context.TODO(), client.ObjectKeyFromObject(keySecret), keySecret); err != nil {
		return nil, err
	}

	keyring, ok := keySecret.Data[clientName]
	if !ok {
		return nil, fmt.Errorf("expected ceph client secret %s to have key %s", keySecretName, clientName)
	}

	monLabels := labels.Set{"ceph_daemon_type": "mon"}

	var monServices corev1.ServiceList
	if err := c.List(context.TODO(), &monServices, &client.ListOptions{
		Namespace:     rookNamespace,
		LabelSelector: monLabels.AsSelector(),
	}); err != nil {
		return nil, err
	}

	var monHosts []string
	for _, svc := range monServices.Items {
		monHosts = append(monHosts, svc.Spec.ClusterIP)
	}

	// TODO labels
	secret := template.GenericSecret(name, namespace, nil)

	secret.StringData["ceph.conf"] = template.MustRenderFile(AppLabel, "ceph.conf", clientSecretOptions{
		MonHost:    strings.Join(monHosts, ","),
		ClientName: clientName,
	})
	secret.StringData["keyring"] = template.MustRenderFile(AppLabel, "keyring", clientSecretKeyringOptions{
		ClientName: clientName,
		Keyring:    string(keyring),
	})

	return secret, nil
}

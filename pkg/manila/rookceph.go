package manila

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/rookceph"
)

type cephClientKey struct {
	Name, Namespace string
}

func RookCephResources(instance *openstackv1beta1.Manila) []*unstructured.Unstructured {
	backends := filterRookCephBackends(instance.Spec.Backends)

	clients := make(map[cephClientKey]bool)
	for _, backend := range backends {
		cephSpec := backend.Ceph
		key := cephClientKey{cephSpec.ClientName, cephSpec.Rook.Namespace}
		clients[key] = true
	}

	var resources []*unstructured.Unstructured

	for client := range clients {
		resources = append(resources, rookceph.Client(client.Namespace, rookceph.ClientOptions{
			Name: client.Name,
			Caps: map[string]string{
				"mgr": "allow rw",
				// TODO need to reduce this
				"mon": "allow *",
				"osd": "allow rw",
			},
		}))
	}

	return resources
}

func RookCephSecrets(ctx context.Context, c client.Client, instance *openstackv1beta1.Manila) ([]*corev1.Secret, error) {
	backends := filterRookCephBackends(instance.Spec.Backends)

	namespaces := make(map[string]bool)
	clientsBySecrets := make(map[string]cephClientKey)
	for _, backend := range backends {
		cephSpec := backend.Ceph
		namespaces[cephSpec.Rook.Namespace] = true
		// TODO validate that all backends with this secret name are compatible
		clientsBySecrets[cephSpec.Secret] = cephClientKey{cephSpec.ClientName, cephSpec.Rook.Namespace}
	}

	// collect mon addrs for each Rook namespace
	monsByNamespace := make(map[string][]string)
	for ns := range namespaces {
		addrs, err := rookceph.GetCephMonitorAddrs(ctx, c, ns)
		if err != nil {
			return nil, err
		}
		monsByNamespace[ns] = addrs
	}

	// collect client secrets
	var secrets []*corev1.Secret
	for secretName, client := range clientsBySecrets {
		keyring, err := rookceph.GetCephClientSecret(ctx, c, client.Name, client.Namespace)
		if err != nil {
			return nil, err
		}

		monHosts := monsByNamespace[client.Namespace]

		secrets = append(secrets, rookceph.ClientSecret(secretName, instance.Namespace, client.Name, keyring, monHosts))
	}

	return secrets, nil
}

func filterRookCephBackends(allBackends []openstackv1beta1.ManilaBackendSpec) []openstackv1beta1.ManilaBackendSpec {
	var backends []openstackv1beta1.ManilaBackendSpec
	for _, backend := range allBackends {
		if backend.Ceph == nil || backend.Ceph.Rook == nil {
			continue
		}
		backends = append(backends, backend)
	}
	return backends
}

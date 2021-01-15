package rookceph

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/ianunruh/openstack-operator/pkg/template"
)

type BlockPoolOptions struct {
	Name           string
	FailureDomain  string
	ReplicatedSize string
}

func BlockPool(namespace string, opts BlockPoolOptions) *unstructured.Unstructured {
	manifest := template.MustRenderFile(AppLabel, "ceph-block-pool.yaml", opts)

	res := template.MustDecodeManifest(manifest)
	res.SetNamespace(namespace)

	return res
}

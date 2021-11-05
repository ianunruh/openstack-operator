package prometheus

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/ianunruh/openstack-operator/pkg/template"
)

type ServiceMonitorParams struct {
	Name          string
	Namespace     string
	NameLabel     string
	InstanceLabel string
}

func ServiceMonitor(params ServiceMonitorParams) *unstructured.Unstructured {
	manifest := template.MustRenderFile("prometheus", "servicemonitor.yaml", params)

	res := template.MustDecodeManifest(manifest)
	res.SetNamespace(params.Namespace)

	return res
}

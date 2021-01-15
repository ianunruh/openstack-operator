package glance

import (
	"fmt"
	"strconv"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/rookceph"
)

func RookCephResources(instance *openstackv1beta1.Glance) []*unstructured.Unstructured {
	spec := instance.Spec.Storage.RookCeph

	return []*unstructured.Unstructured{
		rookceph.Client(spec.Namespace, rookceph.ClientOptions{
			Name: spec.ClientName,
			Caps: map[string]string{
				"mon": "profile rbd",
				"osd": fmt.Sprintf("profile rbd pool=%s", spec.PoolName),
			},
		}),
		rookceph.BlockPool(spec.Namespace, rookceph.BlockPoolOptions{
			Name:           spec.PoolName,
			FailureDomain:  "host",
			ReplicatedSize: strconv.Itoa(spec.ReplicatedSize),
		}),
	}
}

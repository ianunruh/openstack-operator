package cinder

import (
	"fmt"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/rookceph"
)

func RookCephResources(instance *openstackv1beta1.Cinder, imagePool string) []*unstructured.Unstructured {
	spec := instance.Spec.Volume.Storage.RookCeph

	caps := []string{
		fmt.Sprintf("profile rbd pool=%s", spec.PoolName),
		fmt.Sprintf("profile rbd-read-only pool=%s", imagePool),
	}

	return []*unstructured.Unstructured{
		rookceph.Client(spec.Namespace, rookceph.ClientOptions{
			Name: spec.ClientName,
			Caps: map[string]string{
				"mon": "profile rbd",
				"osd": strings.Join(caps, " "),
			},
		}),
		rookceph.BlockPool(spec.Namespace, rookceph.BlockPoolOptions{
			Name:           spec.PoolName,
			FailureDomain:  "host",
			ReplicatedSize: strconv.Itoa(spec.ReplicatedSize),
		}),
	}
}

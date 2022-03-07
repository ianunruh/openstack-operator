package computeset

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
	"github.com/ianunruh/openstack-operator/pkg/template"
)

const (
	Label = "openstack.ospk8s.com/nova-compute-set"
)

func Reconcile(ctx context.Context, c client.Client, instance *openstackv1beta1.NovaComputeSet, log logr.Logger) error {
	nodes, err := listNodes(ctx, c, instance)
	if err != nil {
		return err
	}

	seenNodes := make(map[string]bool, len(nodes))
	for _, node := range nodes {
		seenNodes[node.Name] = true
	}

	computeNodes, err := listComputeNodes(ctx, c, instance)
	if err != nil {
		return err
	}

	computeNodesByName := make(map[string]*openstackv1beta1.NovaComputeNode, len(computeNodes))
	for i, computeNode := range computeNodes {
		if !seenNodes[computeNode.Spec.Node] {
			// TODO clean up orphaned compute node
			continue
		}
		computeNodesByName[computeNode.Spec.Node] = &computeNodes[i]
	}

	var computeNodesToUpdate []*openstackv1beta1.NovaComputeNode
	for _, node := range nodes {
		intended := newComputeNode(instance, node)
		controllerutil.SetControllerReference(instance, intended, c.Scheme())

		hash, err := template.ObjectHash(instance)
		if err != nil {
			return fmt.Errorf("error hashing object: %w", err)
		}

		current := computeNodesByName[node.Name]
		if current == nil {
			template.SetAppliedHash(intended, hash)

			log.Info("Creating NovaComputeNode", "Name", intended.Name)
			if err := c.Create(ctx, intended); err != nil {
				return err
			}
		} else if !template.MatchesAppliedHash(current, hash) {
			current.Spec = intended.Spec
			template.SetAppliedHash(current, hash)

			// queue update for later
			computeNodesToUpdate = append(computeNodesToUpdate, current)
		}
	}

	for _, computeNode := range computeNodesToUpdate {
		log.Info("Updating NovaComputeNode", "Name", computeNode.Name)
		if err := c.Update(ctx, computeNode); err != nil {
			return err
		}
		// TODO wait for node to reflect update
	}

	return nil
}

func listNodes(ctx context.Context, c client.Client, instance *openstackv1beta1.NovaComputeSet) ([]corev1.Node, error) {
	opts := &client.ListOptions{}
	if len(instance.Spec.NodeSelector) > 0 {
		opts.LabelSelector = labels.Set(instance.Spec.NodeSelector).AsSelector()
	}

	var result corev1.NodeList
	if err := c.List(ctx, &result, opts); err != nil {
		return nil, err
	}
	return result.Items, nil
}

func listComputeNodes(ctx context.Context, c client.Client, instance *openstackv1beta1.NovaComputeSet) ([]openstackv1beta1.NovaComputeNode, error) {
	opts := &client.ListOptions{
		LabelSelector: labels.Set(map[string]string{
			Label: instance.Name,
		}).AsSelector(),
		Namespace: instance.Namespace,
	}

	var result openstackv1beta1.NovaComputeNodeList
	if err := c.List(ctx, &result, opts); err != nil {
		return nil, err
	}
	return result.Items, nil
}

func newComputeNode(instance *openstackv1beta1.NovaComputeSet, node corev1.Node) *openstackv1beta1.NovaComputeNode {
	return &openstackv1beta1.NovaComputeNode{
		ObjectMeta: metav1.ObjectMeta{
			Name:      template.Combine(instance.Name, node.Name),
			Namespace: instance.Namespace,
			Labels: map[string]string{
				Label: instance.Name,
			},
		},
		Spec: openstackv1beta1.NovaComputeNodeSpec{
			Node:     node.Name,
			Cell:     instance.Spec.Cell,
			Image:    instance.Spec.Image,
			Libvirtd: instance.Spec.Libvirtd,
		},
	}
}

func Ensure(ctx context.Context, c client.Client, instance *openstackv1beta1.NovaComputeSet, log logr.Logger) error {
	intended := instance.DeepCopy()
	hash, err := template.ObjectHash(instance)
	if err != nil {
		return fmt.Errorf("error hashing object: %w", err)
	}

	if err := c.Get(ctx, client.ObjectKeyFromObject(instance), instance); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}

		template.SetAppliedHash(intended, hash)

		log.Info("Creating NovaComputeSet", "Name", instance.Name)
		return c.Create(ctx, instance)
	} else if !template.MatchesAppliedHash(instance, hash) {
		instance.Spec = intended.Spec
		template.SetAppliedHash(instance, hash)

		log.Info("Updating NovaComputeSet", "Name", instance.Name)
		return c.Update(ctx, instance)
	}

	return nil
}

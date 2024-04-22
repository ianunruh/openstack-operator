/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	"errors"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var octavialog = logf.Log.WithName("octavia-resource")

const (
	DefaultAmphoraImageURL = "https://tarballs.opendev.org/openstack/octavia/test-images/test-only-amphora-x64-haproxy-ubuntu-jammy.qcow2"
)

func (r *Octavia) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-openstack-ospk8s-com-v1beta1-octavia,mutating=true,failurePolicy=fail,sideEffects=None,groups=openstack.ospk8s.com,resources=octavias,verbs=create;update,versions=v1beta1,name=moctavia.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &Octavia{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Octavia) Default() {
	octavialog.Info("default", "name", r.Name)

	r.Spec.Broker = brokerDefault(r.Spec.Broker, r.Name, defaultVirtualHost)
	r.Spec.Database = databaseDefault(r.Spec.Database, r.Name)
	r.Spec.Image = imageDefault(r.Spec.Image, OctaviaDefaultImage)

	if r.Spec.Amphora.Enabled {
		if r.Spec.Amphora.ImageURL == "" {
			r.Spec.Amphora.ImageURL = DefaultAmphoraImageURL
		}
		if r.Spec.Amphora.ManagementCIDR == "" {
			r.Spec.Amphora.ManagementCIDR = "172.28.0.0/24"
		}
	}

	r.Spec.API.NodeSelector = nodeSelectorDefault(r.Spec.API.NodeSelector, r.Spec.NodeSelector)
	r.Spec.DriverAgent.NodeSelector = nodeSelectorDefault(r.Spec.DriverAgent.NodeSelector, r.Spec.NodeSelector)
	r.Spec.HealthManager.NodeSelector = nodeSelectorDefault(r.Spec.HealthManager.NodeSelector, r.Spec.NodeSelector)
	r.Spec.Housekeeping.NodeSelector = nodeSelectorDefault(r.Spec.Housekeeping.NodeSelector, r.Spec.NodeSelector)
	r.Spec.Worker.NodeSelector = nodeSelectorDefault(r.Spec.Worker.NodeSelector, r.Spec.NodeSelector)
}

//+kubebuilder:webhook:path=/validate-openstack-ospk8s-com-v1beta1-octavia,mutating=false,failurePolicy=fail,sideEffects=None,groups=openstack.ospk8s.com,resources=octavias,verbs=create;update,versions=v1beta1,name=voctavia.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &Octavia{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Octavia) ValidateCreate() (admission.Warnings, error) {
	octavialog.Info("validate create", "name", r.Name)

	if err := r.validateProviders(); err != nil {
		return admission.Warnings{}, err
	}

	return admission.Warnings{}, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Octavia) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	octavialog.Info("validate update", "name", r.Name)

	if err := r.validateProviders(); err != nil {
		return admission.Warnings{}, err
	}

	// TODO amphora managementCIDR should be immutable

	return admission.Warnings{}, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Octavia) ValidateDelete() (admission.Warnings, error) {
	octavialog.Info("validate delete", "name", r.Name)

	return admission.Warnings{}, nil
}

func (r *Octavia) validateProviders() error {
	if !r.Spec.Amphora.Enabled && !r.Spec.OVN.Enabled {
		return errors.New("at least one provider must be enabled")
	}
	return nil
}

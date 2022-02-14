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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var octavialog = logf.Log.WithName("octavia-resource")

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

	if r.Spec.Amphora.ManagementCIDR == "" {
		r.Spec.Amphora.ManagementCIDR = "172.28.0.0/24"
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
func (r *Octavia) ValidateCreate() error {
	octavialog.Info("validate create", "name", r.Name)

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Octavia) ValidateUpdate(old runtime.Object) error {
	octavialog.Info("validate update", "name", r.Name)

	// TODO amphora managementCIDR should be immutable

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Octavia) ValidateDelete() error {
	octavialog.Info("validate delete", "name", r.Name)

	return nil
}

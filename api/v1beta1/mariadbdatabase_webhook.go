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
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var mariadbdatabaselog = logf.Log.WithName("mariadbdatabase-resource")

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *MariaDBDatabase) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-openstack-ospk8s-com-v1beta1-mariadbdatabase,mutating=true,failurePolicy=fail,sideEffects=None,groups=openstack.ospk8s.com,resources=mariadbdatabases,verbs=create;update,versions=v1beta1,name=mmariadbdatabase.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &MariaDBDatabase{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *MariaDBDatabase) Default() {
	mariadbdatabaselog.Info("default", "name", r.Name)

	r.Spec.SetupJob.Image = imageDefault(r.Spec.SetupJob.Image, DefaultMariaDBImage)
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-openstack-ospk8s-com-v1beta1-mariadbdatabase,mutating=false,failurePolicy=fail,sideEffects=None,groups=openstack.ospk8s.com,resources=mariadbdatabases,verbs=create;update,versions=v1beta1,name=vmariadbdatabase.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &MariaDBDatabase{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *MariaDBDatabase) ValidateCreate() (admission.Warnings, error) {
	mariadbdatabaselog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *MariaDBDatabase) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	mariadbdatabaselog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *MariaDBDatabase) ValidateDelete() (admission.Warnings, error) {
	mariadbdatabaselog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}

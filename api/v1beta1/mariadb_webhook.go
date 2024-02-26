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
var mariadblog = logf.Log.WithName("mariadb-resource")

func (r *MariaDB) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-openstack-ospk8s-com-v1beta1-mariadb,mutating=true,failurePolicy=fail,sideEffects=None,groups=openstack.ospk8s.com,resources=mariadbs,verbs=create;update,versions=v1beta1,name=mmariadb.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &MariaDB{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *MariaDB) Default() {
	mariadblog.Info("default", "name", r.Name)

	r.Spec.Image = imageDefault(r.Spec.Image, MariaDBDefaultImage)
	r.Spec.Volume = volumeDefault(r.Spec.Volume)

	if promSpec := r.Spec.Prometheus; promSpec != nil {
		promSpec.Exporter.Image = imageDefault(promSpec.Exporter.Image, MariaDBExporterDefaultImage)
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-openstack-ospk8s-com-v1beta1-mariadb,mutating=false,failurePolicy=fail,sideEffects=None,groups=openstack.ospk8s.com,resources=mariadbs,verbs=create;update,versions=v1beta1,name=vmariadb.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &MariaDB{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *MariaDB) ValidateCreate() (admission.Warnings, error) {
	mariadblog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return admission.Warnings{}, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *MariaDB) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	mariadblog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return admission.Warnings{}, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *MariaDB) ValidateDelete() (admission.Warnings, error) {
	mariadblog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return admission.Warnings{}, nil
}

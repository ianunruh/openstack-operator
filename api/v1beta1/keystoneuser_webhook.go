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

var keystoneuserlog = logf.Log.WithName("keystoneuser-resource")

func (r *KeystoneUser) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-openstack-ospk8s-com-v1beta1-keystoneuser,mutating=true,failurePolicy=fail,sideEffects=None,groups=openstack.ospk8s.com,resources=keystoneusers,verbs=create;update,versions=v1beta1,name=mkeystoneuser.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &KeystoneUser{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *KeystoneUser) Default() {
	keystoneuserlog.Info("default", "name", r.Name)

	if r.Spec.Domain == "" {
		r.Spec.Domain = "Default"
	}

	if r.Spec.ProjectDomain == "" {
		r.Spec.ProjectDomain = "Default"
	}

	if len(r.Spec.Roles) == 0 {
		r.Spec.Roles = []string{"admin"}
	}
}

//+kubebuilder:webhook:path=/validate-openstack-ospk8s-com-v1beta1-keystoneuser,mutating=false,failurePolicy=fail,sideEffects=None,groups=openstack.ospk8s.com,resources=keystoneusers,verbs=create;update,versions=v1beta1,name=vkeystoneuser.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &KeystoneUser{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *KeystoneUser) ValidateCreate() (admission.Warnings, error) {
	keystoneuserlog.Info("validate create", "name", r.Name)
	return admission.Warnings{}, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *KeystoneUser) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	keystoneuserlog.Info("validate update", "name", r.Name)
	return admission.Warnings{}, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *KeystoneUser) ValidateDelete() (admission.Warnings, error) {
	keystoneuserlog.Info("validate delete", "name", r.Name)
	return admission.Warnings{}, nil
}

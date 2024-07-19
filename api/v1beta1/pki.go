package v1beta1

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

type TLSClientSpec struct {
	// +optional
	CABundle string `json:"caBundle,omitempty"`
}

type TLSProxySpec struct {
	// +optional
	Image string `json:"image,omitempty"`

	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type TLSServerSpec struct {
	// +optional
	Secret string `json:"secret,omitempty"`

	// +optional
	Issuer IssuerRef `json:"issuer,omitempty"`
}

type IssuerRef struct {
	// +optional
	Name string `json:"name,omitempty"`

	// +optional
	Kind string `json:"kind,omitempty"`
}

func tlsProxyDefault(spec TLSProxySpec) TLSProxySpec {
	spec.Image = imageDefault(spec.Image, DefaultTLSProxyImage)
	return spec
}

func tlsServerDefault(spec TLSServerSpec, name ...string) TLSServerSpec {
	if spec.Secret == "" {
		if spec.Issuer.Name != "" {
			spec.Secret = fmt.Sprintf("%s-tls", strings.Join(name, "-"))
		}
	}
	return spec
}

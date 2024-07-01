package v1beta1

import "fmt"

type TLSClientSpec struct {
	// +optional
	CASecrets []string `json:"caSecrets,omitempty"`
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

func tlsServerDefault(spec TLSServerSpec, name, component string) TLSServerSpec {
	if spec.Secret == "" {
		if spec.Issuer.Name != "" {
			spec.Secret = fmt.Sprintf("%s-%s", name, component)
		}
	}
	return spec
}

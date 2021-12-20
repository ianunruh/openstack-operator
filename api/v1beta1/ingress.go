package v1beta1

type IngressSpec struct {
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// +optional
	ClassName *string `json:"className,omitempty"`

	// +optional
	Host string `json:"host,omitempty"`

	// +optional
	TLSSecretName string `json:"tlsSecretName,omitempty"`
}

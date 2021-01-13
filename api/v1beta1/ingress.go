package v1beta1

type IngressSpec struct {
	// +optional
	Host string `json:"host"`

	// +optional
	Annotations map[string]string `json:"annotations"`
}

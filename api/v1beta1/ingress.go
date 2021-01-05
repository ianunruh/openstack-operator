package v1beta1

type IngressSpec struct {
	Host string `json:"host"`

	Annotations map[string]string `json:"annotations"`
}

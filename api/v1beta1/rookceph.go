package v1beta1

type RookCephSpec struct {
	// +optional
	Namespace string `json:"namespace"`

	// +optional
	PoolName string `json:"poolName"`

	// +optional
	ClientName string `json:"clientName"`

	// +optional
	Secret string `json:"secret"`

	// +optional
	DeviceClass string `json:"deviceClass"`

	ReplicatedSize int `json:"replicatedSize"`
}

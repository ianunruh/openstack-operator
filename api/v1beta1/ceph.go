package v1beta1

type CephSpec struct {
	PoolName string `json:"poolName"`

	ClientName string `json:"clientName"`

	Secret string `json:"secret"`

	Rook *RookCephSpec `json:"rook"`
}

type RookCephSpec struct {
	Namespace string `json:"namespace"`

	// +optional
	DeviceClass string `json:"deviceClass"`

	ReplicatedSize int `json:"replicatedSize"`
}

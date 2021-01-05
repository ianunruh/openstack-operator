package v1beta1

type VolumeSpec struct {
	Capacity     string  `json:"capacity"`
	StorageClass *string `json:"storageClass,omitempty"`
}

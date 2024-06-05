package v1beta1

const (
	DefaultCacheHost        = "memcached"
	DefaultCachePort uint16 = 11211
)

type CacheSpec struct {
	// +optional
	Host string `json:"host"`

	// +optional
	Port uint16 `json:"port,omitempty"`
}

func cacheDefault(spec CacheSpec) CacheSpec {
	if spec.Host == "" {
		spec.Host = DefaultCacheHost
	}

	if spec.Port == 0 {
		spec.Port = DefaultCachePort
	}

	return spec
}

package v1beta1

const DefaultCacheServer = "memcached:11211"

type CacheSpec struct {
	// +optional
	Servers []string `json:"servers,omitempty"`
}

func cacheDefault(spec CacheSpec) CacheSpec {
	if spec.Servers == nil {
		spec.Servers = []string{DefaultCacheServer}
	}

	return spec
}

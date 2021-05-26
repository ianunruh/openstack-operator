package template

import "strings"

const (
	AppLabel       = "app.kubernetes.io/name"
	ComponentLabel = "app.kubernetes.io/component"
	InstanceLabel  = "app.kubernetes.io/instance"
)

func Labels(instance, app, component string) map[string]string {
	return map[string]string{
		AppLabel:       app,
		InstanceLabel:  instance,
		ComponentLabel: component,
	}
}

func AppLabels(instance, app string) map[string]string {
	return map[string]string{
		AppLabel:      app,
		InstanceLabel: instance,
	}
}

func Combine(parts ...string) string {
	return strings.Join(parts, "-")
}

func MergeStringMaps(maps ...map[string]string) map[string]string {
	var merged map[string]string
	for _, m := range maps {
		if merged == nil {
			merged = make(map[string]string, len(m))
		}
		for k, v := range m {
			merged[k] = v
		}
	}

	return merged
}

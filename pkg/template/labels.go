package template

import "strings"

const (
	AppLabel       = "app"
	InstanceLabel  = "instance"
	ComponentLabel = "component"
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

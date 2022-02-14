package v1beta1

func nodeSelectorDefault(selector, defaultSelector map[string]string) map[string]string {
	if selector == nil {
		return defaultSelector
	}
	return selector
}

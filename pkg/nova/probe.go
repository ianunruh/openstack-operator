package nova

import (
	"strconv"

	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

const (
	amqpPort uint16 = 5672
)

func amqpHealthProbeHandler(processName string, brokerSpec openstackv1beta1.RabbitMQUserSpec) corev1.ProbeHandler {
	port := amqpPort
	if externalSpec := brokerSpec.External; externalSpec != nil {
		if externalSpec.Port > 0 {
			port = externalSpec.Port
		}
	}
	return corev1.ProbeHandler{
		Exec: &corev1.ExecAction{
			Command: []string{"healthcheck_port", processName, strconv.Itoa(int(port))},
		},
	}
}

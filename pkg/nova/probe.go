package nova

import (
	"strconv"

	corev1 "k8s.io/api/core/v1"

	openstackv1beta1 "github.com/ianunruh/openstack-operator/api/v1beta1"
)

const (
	amqpPort    uint16 = 5672
	amqpTLSPort uint16 = 5671
)

func amqpHealthProbeHandler(processName string, brokerSpec openstackv1beta1.RabbitMQUserSpec) corev1.ProbeHandler {
	port := amqpPort
	if externalSpec := brokerSpec.External; externalSpec != nil {
		if externalSpec.Port > 0 {
			port = externalSpec.Port
		} else if externalSpec.TLS.CABundle != "" {
			port = amqpTLSPort
		}
	} else if brokerSpec.TLS.CABundle != "" {
		port = amqpTLSPort
	}
	return corev1.ProbeHandler{
		Exec: &corev1.ExecAction{
			Command: []string{"healthcheck_port", processName, strconv.Itoa(int(port))},
		},
	}
}

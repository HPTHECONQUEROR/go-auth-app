package usecase

import (
	"encoding/json"
	"fmt"
	"go-auth-app/pkg"
	"time"
)

type NATSUsecase struct {
	NATSClient *pkg.NatsClient
}

func NewNATSUsecase(natsClient *pkg.NatsClient) *NATSUsecase {
	return &NATSUsecase{NATSClient: natsClient}
}

func (u *NATSUsecase) GetTopics() ([]string, error) {
	// This is a placeholder. Actual implementation depends on NATS client capabilities
	return []string{"monitoring.snmp"}, nil
}

func (u *NATSUsecase) PublishMessage(topic, message string) error {
	return u.NATSClient.Publish(topic, []byte(message))
}

func (u *NATSUsecase) SimulateSNMPMetrics() (map[string]interface{}, error) {
	metrics := map[string]interface{}{
		"device_id":    "simulator_001",
		"timestamp":    time.Now(),
		"cpu_load":     50,
		"memory_usage": 65.5,
		"network": map[string]interface{}{
			"in_octets":  1024,
			"out_octets": 512,
		},
	}

	// Publish to NATS
	jsonMetrics, err := json.Marshal(metrics)
	if err != nil {
		return nil, err
	}

	err = u.NATSClient.Publish("monitoring.snmp", jsonMetrics)
	if err != nil {
		return nil, fmt.Errorf("failed to publish metrics: %v", err)
	}

	return metrics, nil
}
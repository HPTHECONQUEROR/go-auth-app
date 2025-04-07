package usecase

import (
	"encoding/json"
	"fmt"
	"go-auth-app/pkg"
	"sync"
	"time"
	
	"github.com/nats-io/nats.go"
)

type NATSUsecase struct {
	NATSClient *pkg.NatsClient
	// Store the latest metrics for each topic
	metrics     map[string]interface{}
	metricsMutex sync.RWMutex
}

func NewNATSUsecase(natsClient *pkg.NatsClient) *NATSUsecase {
	usecase := &NATSUsecase{
		NATSClient: natsClient,
		metrics:    make(map[string]interface{}),
	}
	
	// Start the subscription to SNMP metrics
	usecase.SetupSubscriptions()
	
	return usecase
}

// SetupSubscriptions initializes all needed NATS subscriptions
func (u *NATSUsecase) SetupSubscriptions() {
	// Subscribe to the SNMP metrics topic
	if u.NATSClient.IsConnected() {
		_, err := u.NATSClient.Subscribe("monitoring.snmp", func(msg *nats.Msg) {
			var metrics map[string]interface{}
			if err := json.Unmarshal(msg.Data, &metrics); err != nil {
				fmt.Printf("Error unmarshaling metrics: %v\n", err)
				return
			}
			
			// Store the metrics
			u.metricsMutex.Lock()
			u.metrics["monitoring.snmp"] = metrics
			u.metricsMutex.Unlock()
			
			fmt.Printf("Received metrics on topic %s\n", msg.Subject)
		})
		
		if err != nil {
			fmt.Printf("Failed to subscribe to monitoring.snmp: %v\n", err)
		} else {
			fmt.Println("Successfully subscribed to monitoring.snmp")
		}
	}
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
		"device_id":      "pi4-iot-node-007",
		"timestamp":      time.Now(),
		"device_name":    "RaspberryPi4ModelB",
		"cpu_load":       50,
		"memory_usage":   65.5,
		"memory_total":   8192,
		"in_octets":      1024,
		"out_octets":     512,
		"in_errors":      0,
		"out_errors":     0,
		"device_uptime":  "10d 5h 30m",
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

// GetLatestMetrics returns the latest metrics for a given topic
func (u *NATSUsecase) GetLatestMetrics(topic string) (interface{}, error) {
	u.metricsMutex.RLock()
	defer u.metricsMutex.RUnlock()
	
	metrics, exists := u.metrics[topic]
	if !exists {
		return nil, fmt.Errorf("no metrics available for topic: %s", topic)
	}
	
	return metrics, nil
}
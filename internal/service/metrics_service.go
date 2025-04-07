package service

import (
	"encoding/json"
	"go-auth-app/internal/domain"
	"go-auth-app/internal/repository"
	"go-auth-app/pkg"
	"log"
	"sync"
	
	"github.com/nats-io/nats.go"
)

// MetricsService handles storing and retrieving metrics data
type MetricsService struct {
	NatsClient     *pkg.NatsClient
	MetricsRepo    repository.MetricsRepository
	Topic          string
	subscription   *nats.Subscription
	lastMetrics    *domain.SNMPMetrics
	metricsMutex   sync.RWMutex
}

// NewMetricsService creates a new metrics service
func NewMetricsService(natsClient *pkg.NatsClient, metricsRepo repository.MetricsRepository, topic string) *MetricsService {
	return &MetricsService{
		NatsClient:  natsClient,
		MetricsRepo: metricsRepo,
		Topic:       topic,
	}
}

// Start begins listening for metrics
func (s *MetricsService) Start() error {
	// Subscribe to the NATS topic
	subscription, err := s.NatsClient.Conn.Subscribe(s.Topic, s.handleMetricsMessage)
	if err != nil {
		return err
	}

	s.subscription = subscription
	log.Printf("Metrics service started, listening on topic: %s", s.Topic)
	return nil
}

// Stop stops the metrics service
func (s *MetricsService) Stop() error {
	if s.subscription != nil {
		err := s.subscription.Unsubscribe()
		if err != nil {
			return err
		}
	}

	log.Println("Metrics service stopped")
	return nil
}

// GetLatestMetrics returns the most recent metrics
func (s *MetricsService) GetLatestMetrics() *domain.SNMPMetrics {
	s.metricsMutex.RLock()
	defer s.metricsMutex.RUnlock()

	if s.lastMetrics == nil {
		return nil
	}

	// Return a copy to avoid concurrent access issues
	metricsCopy := *s.lastMetrics
	return &metricsCopy
}

// handleMetricsMessage processes incoming metrics messages from NATS
func (s *MetricsService) handleMetricsMessage(msg *nats.Msg) {
	var metrics domain.SNMPMetrics
	if err := json.Unmarshal(msg.Data, &metrics); err != nil {
		log.Printf("Error unmarshaling metrics: %v", err)
		return
	}

	// Store metrics
	s.metricsMutex.Lock()
	s.lastMetrics = &metrics
	s.metricsMutex.Unlock()

	// Store in repository
	err := s.MetricsRepo.StoreMetrics(&metrics)
	if err != nil {
		log.Printf("Error storing metrics: %v", err)
	}

	log.Printf("Received and stored metrics from device: %s", metrics.DeviceInfo.Name)
}
package snmp

import (
	"context"
	"encoding/json"
	"fmt"
	"go-auth-app/internal/domain"
	"go-auth-app/pkg"
	"log"
	"sync"
	"time"

	"github.com/gosnmp/gosnmp"
)

// SNMPCollector collects metrics from an SNMP agent
type SNMPCollector struct {
	Host          string
	Port          uint16
	Community     string
	Version       gosnmp.SnmpVersion
	Timeout       time.Duration
	Interval      time.Duration
	NatsClient    *pkg.NatsClient
	Topic         string
	stopCh        chan struct{}
	wg            sync.WaitGroup
	isRunning     bool
	runningMutex  sync.Mutex
	lastMetrics   *domain.SNMPMetrics
	metricsMutex  sync.RWMutex
}

// NewSNMPCollector creates a new SNMP collector
func NewSNMPCollector(host string, port uint16, community string, natsClient *pkg.NatsClient, topic string) *SNMPCollector {
	return &SNMPCollector{
		Host:       host,
		Port:       port,
		Community:  community,
		Version:    gosnmp.Version2c,
		Timeout:    5 * time.Second,
		Interval:   30 * time.Second,
		NatsClient: natsClient,
		Topic:      topic,
		stopCh:     make(chan struct{}),
	}
}

// Start begins the SNMP collection
func (c *SNMPCollector) Start(ctx context.Context) error {
	c.runningMutex.Lock()
	defer c.runningMutex.Unlock()

	if c.isRunning {
		return fmt.Errorf("collector is already running")
	}

	c.isRunning = true
	c.stopCh = make(chan struct{})

	// Start the collection loop
	c.wg.Add(1)
	go c.collectLoop(ctx)

	log.Printf("SNMP Collector started for %s:%d", c.Host, c.Port)
	return nil
}

// Stop stops the SNMP collection
func (c *SNMPCollector) Stop() {
	c.runningMutex.Lock()
	defer c.runningMutex.Unlock()

	if !c.isRunning {
		return
	}

	close(c.stopCh)
	c.wg.Wait()
	c.isRunning = false

	log.Printf("SNMP Collector stopped for %s:%d", c.Host, c.Port)
}

// GetLastMetrics returns the last collected metrics
func (c *SNMPCollector) GetLastMetrics() *domain.SNMPMetrics {
	c.metricsMutex.RLock()
	defer c.metricsMutex.RUnlock()
	
	if c.lastMetrics == nil {
		return nil
	}
	
	// Return a copy to avoid concurrent access issues
	metricsCopy := *c.lastMetrics
	return &metricsCopy
}

// collectLoop periodically collects SNMP metrics
func (c *SNMPCollector) collectLoop(ctx context.Context) {
	defer c.wg.Done()

	ticker := time.NewTicker(c.Interval)
	defer ticker.Stop()

	// Collect immediately on start
	c.collect()

	for {
		select {
		case <-ticker.C:
			c.collect()
		case <-c.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// collect performs a single SNMP collection
func (c *SNMPCollector) collect() {
	metrics, err := c.fetchSNMPMetrics()
	if err != nil {
		log.Printf("Error collecting SNMP metrics: %v", err)
		return
	}

	// Store the metrics
	c.metricsMutex.Lock()
	c.lastMetrics = metrics
	c.metricsMutex.Unlock()

	// Publish to NATS
	c.publishMetrics(metrics)
}

// fetchSNMPMetrics collects metrics from the SNMP agent
func (c *SNMPCollector) fetchSNMPMetrics() (*domain.SNMPMetrics, error) {
	client := &gosnmp.GoSNMP{
		Target:    c.Host,
		Port:      c.Port,
		Community: c.Community,
		Version:   c.Version,
		Timeout:   c.Timeout,
		Retries:   3,
	}

	err := client.Connect()
	if err != nil {
		return nil, fmt.Errorf("connect error: %v", err)
	}
	defer client.Conn.Close()

	// OIDs for common metrics
	oids := []string{
		".1.3.6.1.2.1.1.1.0",  // sysDescr
		".1.3.6.1.2.1.1.3.0",  // sysUpTime
		".1.3.6.1.2.1.1.5.0",  // sysName
		".1.3.6.1.2.1.25.2.2.0", // hrMemorySize
		".1.3.6.1.2.1.25.3.3.1.2.1", // hrProcessorLoad
	}

	// Fetch the metrics
	result, err := client.Get(oids)
	if err != nil {
		return nil, fmt.Errorf("get error: %v", err)
	}

	// Process results
	metrics := &domain.SNMPMetrics{
		Timestamp: time.Now(),
		DeviceInfo: domain.DeviceInfo{
			Name:         "Unknown",
			Description:  "Unknown",
			UpTime:       "Unknown",
		},
		SystemMetrics: domain.SystemMetrics{
			CPULoad:      0,
			MemoryTotal:  0,
		},
	}

	// Extract values from the result
	for _, variable := range result.Variables {
		switch variable.Name {
		case ".1.3.6.1.2.1.1.1.0": // sysDescr
			metrics.DeviceInfo.Description = string(variable.Value.([]byte))
		case ".1.3.6.1.2.1.1.3.0": // sysUpTime
			metrics.DeviceInfo.UpTime = fmt.Sprintf("%v", variable.Value)
		case ".1.3.6.1.2.1.1.5.0": // sysName
			metrics.DeviceInfo.Name = string(variable.Value.([]byte))
		case ".1.3.6.1.2.1.25.2.2.0": // hrMemorySize
			metrics.SystemMetrics.MemoryTotal = variable.Value.(uint)
		case ".1.3.6.1.2.1.25.3.3.1.2.1": // hrProcessorLoad
			metrics.SystemMetrics.CPULoad = variable.Value.(int)
		}
	}

	return metrics, nil
}

// publishMetrics publishes metrics to NATS
func (c *SNMPCollector) publishMetrics(metrics *domain.SNMPMetrics) {
	data, err := json.Marshal(metrics)
	if err != nil {
		log.Printf("Error marshaling metrics: %v", err)
		return
	}

	err = c.NatsClient.Conn.Publish(c.Topic, data)
	if err != nil {
		log.Printf("Error publishing metrics to NATS: %v", err)
		return
	}

	log.Printf("Published SNMP metrics to NATS topic %s", c.Topic)
}
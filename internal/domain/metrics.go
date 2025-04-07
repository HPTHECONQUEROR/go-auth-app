package domain

import (
	"time"
)

// DeviceInfo contains basic information about a monitored device
type DeviceInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UpTime      string `json:"uptime"`
}

// SystemMetrics contains system metrics from SNMP
type SystemMetrics struct {
	CPULoad     int  `json:"cpu_load"`     // Percentage of CPU utilization
	MemoryTotal uint `json:"memory_total"` // Total memory in KB
}

// NetworkMetrics contains network interface metrics
type NetworkMetrics struct {
	InOctets    uint64 `json:"in_octets"`     // Incoming bytes
	OutOctets   uint64 `json:"out_octets"`    // Outgoing bytes
	InErrors    uint   `json:"in_errors"`     // Incoming errors
	OutErrors   uint   `json:"out_errors"`    // Outgoing errors
	InDiscards  uint   `json:"in_discards"`   // Incoming packets discarded
	OutDiscards uint   `json:"out_discards"`  // Outgoing packets discarded
}

// SNMPMetrics holds all the metrics collected from an SNMP agent
type SNMPMetrics struct {
	DeviceID      string         `json:"device_id"`
	Timestamp     time.Time      `json:"timestamp"`
	DeviceInfo    DeviceInfo     `json:"device_info"`
	SystemMetrics SystemMetrics  `json:"system_metrics"`
	Network       NetworkMetrics `json:"network,omitempty"`
}

// MetricsResponse is the API response structure for metrics data
type MetricsResponse struct {
	Success   bool         `json:"success"`
	Timestamp time.Time    `json:"timestamp"`
	Metrics   *SNMPMetrics `json:"metrics,omitempty"`
	Error     string       `json:"error,omitempty"`
}
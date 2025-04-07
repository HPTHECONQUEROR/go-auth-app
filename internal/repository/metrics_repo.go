package repository

import (
	"context"
	"fmt"
	"go-auth-app/db"
	"go-auth-app/internal/domain"
	"sync"
	"time"
)

// MetricsRepository defines the interface for metrics data storage
type MetricsRepository interface {
	StoreMetrics(metrics *domain.SNMPMetrics) error
	GetLatestMetrics() (*domain.SNMPMetrics, error)
	GetMetricsByTimeRange(start, end time.Time) ([]*domain.SNMPMetrics, error)
}

// metricsRepo implements MetricsRepository with in-memory and DB storage
type metricsRepo struct {
	maxInMemory     int
	recentMetrics   []*domain.SNMPMetrics
	metricsLock     sync.RWMutex
}

// NewMetricsRepository creates a new metrics repository
func NewMetricsRepository(maxInMemory int) MetricsRepository {
	return &metricsRepo{
		maxInMemory:   maxInMemory,
		recentMetrics: make([]*domain.SNMPMetrics, 0, maxInMemory),
	}
}

// StoreMetrics stores metrics in memory and database
func (r *metricsRepo) StoreMetrics(metrics *domain.SNMPMetrics) error {
	// Store in memory
	r.metricsLock.Lock()
	r.recentMetrics = append(r.recentMetrics, metrics)
	if len(r.recentMetrics) > r.maxInMemory {
		// Keep only the most recent metrics
		r.recentMetrics = r.recentMetrics[1:]
	}
	r.metricsLock.Unlock()

	// Store in database
	return r.storeMetricsInDB(metrics)
}

// GetLatestMetrics returns the most recent metrics
func (r *metricsRepo) GetLatestMetrics() (*domain.SNMPMetrics, error) {
	r.metricsLock.RLock()
	defer r.metricsLock.RUnlock()

	if len(r.recentMetrics) == 0 {
		return nil, fmt.Errorf("no metrics available")
	}

	return r.recentMetrics[len(r.recentMetrics)-1], nil
}

// GetMetricsByTimeRange retrieves metrics within the specified time range
func (r *metricsRepo) GetMetricsByTimeRange(start, end time.Time) ([]*domain.SNMPMetrics, error) {
	// Attempt to get from memory first
	r.metricsLock.RLock()
	var inMemoryMetrics []*domain.SNMPMetrics
	for _, m := range r.recentMetrics {
		if (m.Timestamp.Equal(start) || m.Timestamp.After(start)) && 
		   (m.Timestamp.Equal(end) || m.Timestamp.Before(end)) {
			inMemoryMetrics = append(inMemoryMetrics, m)
		}
	}
	r.metricsLock.RUnlock()

	// If we have data in memory, return it
	if len(inMemoryMetrics) > 0 {
		return inMemoryMetrics, nil
	}

	// Otherwise query the database
	return r.getMetricsFromDB(start, end)
}

// storeMetricsInDB stores metrics in the database
func (r *metricsRepo) storeMetricsInDB(metrics *domain.SNMPMetrics) error {
	query := `
		INSERT INTO snmp_metrics (
			device_id, timestamp, device_name, device_description, device_uptime,
			cpu_load, memory_total, in_octets, out_octets, in_errors, out_errors
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.DB.Exec(ctx, query,
		metrics.DeviceID,
		metrics.Timestamp,
		metrics.DeviceInfo.Name,
		metrics.DeviceInfo.Description,
		metrics.DeviceInfo.UpTime,
		metrics.SystemMetrics.CPULoad,
		metrics.SystemMetrics.MemoryTotal,
		metrics.Network.InOctets,
		metrics.Network.OutOctets,
		metrics.Network.InErrors,
		metrics.Network.OutErrors,
	)

	return err
}

// getMetricsFromDB retrieves metrics from the database within a time range
func (r *metricsRepo) getMetricsFromDB(start, end time.Time) ([]*domain.SNMPMetrics, error) {
	query := `
		SELECT 
			device_id, timestamp, device_name, device_description, device_uptime,
			cpu_load, memory_total, in_octets, out_octets, in_errors, out_errors
		FROM snmp_metrics
		WHERE timestamp BETWEEN $1 AND $2
		ORDER BY timestamp ASC
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.DB.Query(ctx, query, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []*domain.SNMPMetrics
	for rows.Next() {
		m := &domain.SNMPMetrics{
			DeviceInfo:    domain.DeviceInfo{},
			SystemMetrics: domain.SystemMetrics{},
			Network:       domain.NetworkMetrics{},
		}

		err := rows.Scan(
			&m.DeviceID,
			&m.Timestamp,
			&m.DeviceInfo.Name,
			&m.DeviceInfo.Description,
			&m.DeviceInfo.UpTime,
			&m.SystemMetrics.CPULoad,
			&m.SystemMetrics.MemoryTotal,
			&m.Network.InOctets,
			&m.Network.OutOctets,
			&m.Network.InErrors,
			&m.Network.OutErrors,
		)
		if err != nil {
			return nil, err
		}

		metrics = append(metrics, m)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return metrics, nil
}
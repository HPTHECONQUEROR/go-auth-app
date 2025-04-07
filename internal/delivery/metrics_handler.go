package delivery

import (
	"go-auth-app/internal/domain"
	"go-auth-app/internal/service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// MetricsHandler handles HTTP requests for metrics data
type MetricsHandler struct {
	MetricsService *service.MetricsService
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(metricsService *service.MetricsService) *MetricsHandler {
	return &MetricsHandler{
		MetricsService: metricsService,
	}
}

// GetLatestMetricsHandler returns the most recent metrics
func (h *MetricsHandler) GetLatestMetricsHandler(c *gin.Context) {
	metrics := h.MetricsService.GetLatestMetrics()
	if metrics == nil {
		c.JSON(http.StatusOK, domain.MetricsResponse{
			Success:   false,
			Timestamp: time.Now(),
			Error:     "No metrics available",
		})
		return
	}

	c.JSON(http.StatusOK, domain.MetricsResponse{
		Success:   true,
		Timestamp: time.Now(),
		Metrics:   metrics,
	})
}

// GetMetricsByRangeHandler gets metrics within a time range
func (h *MetricsHandler) GetMetricsByRangeHandler(c *gin.Context) {
	// Default to last 24 hours if not specified
	startStr := c.DefaultQuery("start", time.Now().Add(-24*time.Hour).Format(time.RFC3339))
	endStr := c.DefaultQuery("end", time.Now().Format(time.RFC3339))

	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid start time format. Use RFC3339 format.",
		})
		return
	}

	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid end time format. Use RFC3339 format.",
		})
		return
	}

	// Get metrics from repository
	metrics, err := h.MetricsService.MetricsRepo.GetMetricsByTimeRange(start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if len(metrics) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success":   true,
			"timestamp": time.Now(),
			"metrics":   []interface{}{},
			"message":   "No metrics found in the specified time range",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"timestamp": time.Now(),
		"metrics":   metrics,
		"count":     len(metrics),
	})
}

// GetMetricsSummaryHandler provides a summary of system metrics
func (h *MetricsHandler) GetMetricsSummaryHandler(c *gin.Context) {
	latestMetrics := h.MetricsService.GetLatestMetrics()
	if latestMetrics == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   "No metrics available",
		})
		return
	}

	// Create a summary response with the most important metrics
	summary := gin.H{
		"success":          true,
		"timestamp":        time.Now(),
		"device_name":      latestMetrics.DeviceInfo.Name,
		"device_uptime":    latestMetrics.DeviceInfo.UpTime,
		"cpu_load_percent": latestMetrics.SystemMetrics.CPULoad,
		"memory_total_kb":  latestMetrics.SystemMetrics.MemoryTotal,
		"last_updated":     latestMetrics.Timestamp,
	}

	// Add network metrics if available
	if latestMetrics.Network.InOctets > 0 || latestMetrics.Network.OutOctets > 0 {
		summary["network"] = gin.H{
			"in_octets":    latestMetrics.Network.InOctets,
			"out_octets":   latestMetrics.Network.OutOctets,
			"in_errors":    latestMetrics.Network.InErrors,
			"out_errors":   latestMetrics.Network.OutErrors,
			"in_discards":  latestMetrics.Network.InDiscards,
			"out_discards": latestMetrics.Network.OutDiscards,
		}
	}

	c.JSON(http.StatusOK, summary)
}
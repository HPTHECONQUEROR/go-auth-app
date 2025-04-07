package delivery

import (
	"go-auth-app/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

type NATSHandler struct {
	NATSUsecase *usecase.NATSUsecase
}

func NewNATSHandler(natsUsecase *usecase.NATSUsecase) *NATSHandler {
	return &NATSHandler{NATSUsecase: natsUsecase}
}

func (h *NATSHandler) GetTopicsHandler(c *gin.Context) {
	topics, err := h.NATSUsecase.GetTopics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"topics": topics})
}

func (h *NATSHandler) PublishTestHandler(c *gin.Context) {
	var payload struct {
		Topic   string `json:"topic"`
		Message string `json:"message"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.NATSUsecase.PublishMessage(payload.Topic, payload.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message published successfully"})
}

func (h *NATSHandler) SimulateSNMPMetricsHandler(c *gin.Context) {
	metrics, err := h.NATSUsecase.SimulateSNMPMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"metrics": metrics})
}

// GetTopicMetricsHandler returns the latest metrics for a specific topic
func (h *NATSHandler) GetTopicMetricsHandler(c *gin.Context) {
	topic := c.Param("topic")
	if topic == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "topic parameter is required"})
		return
	}
	
	metrics, err := h.NATSUsecase.GetLatestMetrics(topic)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"topic": topic, "metrics": metrics})
}
package routes

import (
	"github.com/gin-gonic/gin"
	"go-auth-app/internal/delivery"
)

// SetupRoutes defines API routes
func SetupRoutes(router *gin.Engine, authHandler *delivery.AuthHandler, chatHandler *delivery.ChatHandler, wsHandler *delivery.WebSocketHandler, metricsHandler *delivery.MetricsHandler) {
	// Public Routes
	router.POST("/signup", authHandler.SignupHandler)
	router.POST("/login", authHandler.LoginHandler)

	// Protected Routes
	router.GET("/protected", delivery.AuthMiddleware(), authHandler.ProtectedHandler)

	// Chat Routes - All require authentication
	chat := router.Group("/chat")
	chat.Use(delivery.AuthMiddleware())
	{
		// Send a message
		chat.POST("/messages", chatHandler.SendMessageHandler)

		// Get messages from a conversation with a specific user
		chat.GET("/messages/:user_id", chatHandler.GetConversationMessagesHandler)

		// Get all user's conversations
		chat.GET("/conversations", chatHandler.GetUserConversationsHandler)

		// WebSocket endpoint for real-time chat
		chat.GET("/ws", wsHandler.HandleWebSocket)
	}

	// Metrics Routes
	metrics := router.Group("/monitoring")
	metrics.Use(delivery.AuthMiddleware())
	{
		// Get latest metrics
		metrics.GET("/metrics", metricsHandler.GetLatestMetricsHandler)
		
		// Get metrics by time range
		metrics.GET("/metrics/range", metricsHandler.GetMetricsByRangeHandler)
		
		// Get metrics summary
		metrics.GET("/metrics/summary", metricsHandler.GetMetricsSummaryHandler)
	}
}
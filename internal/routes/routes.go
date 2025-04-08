package routes

import (
	"github.com/gin-gonic/gin"
	"go-auth-app/internal/delivery"
)

func SetupRoutes(
	router *gin.Engine,
	authHandler *delivery.AuthHandler,
	chatHandler *delivery.ChatHandler,
	wsHandler *delivery.WebSocketHandler,
	natsHandler *delivery.NATSHandler,
) {
	// Existing routes remain the same
	router.POST("/signup", authHandler.SignupHandler)
	router.POST("/login", authHandler.LoginHandler)

	// Chat routes (existing)
	chat := router.Group("/chat")
	chat.Use(delivery.AuthMiddleware())
	{
		chat.POST("/messages", chatHandler.SendMessageHandler)
		chat.GET("/messages/:user_id", chatHandler.GetConversationMessagesHandler)
		chat.GET("/conversations", chatHandler.GetUserConversationsHandler)
		chat.GET("/ws", wsHandler.HandleWebSocket)
	}

	// NATS and SNMP routes
	nats := router.Group("/nats")
	{
		nats.GET("/topics", natsHandler.GetTopicsHandler)
		nats.POST("/publish", natsHandler.PublishTestHandler)
		// New endpoint to subscribe and get metrics from a specific topic
		nats.GET("/subscribe/:topic", natsHandler.GetTopicMetricsHandler)
	}

	snmp := router.Group("/snmp")
	{
		snmp.GET("/metrics", natsHandler.SimulateSNMPMetricsHandler)
	}
}

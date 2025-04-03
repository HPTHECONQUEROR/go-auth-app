package delivery

import (
	"context"
	"go-auth-app/internal/domain"
	"go-auth-app/internal/usecase"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ChatHandler handles HTTP requests related to chat functionality
type ChatHandler struct {
	ChatUsecase *usecase.ChatUsecase
}

// NewChatHandler creates a new instance of ChatHandler
func NewChatHandler(chatUsecase *usecase.ChatUsecase) *ChatHandler {
	return &ChatHandler{ChatUsecase: chatUsecase}
}

// SendMessageHandler handles sending a new message
func (h *ChatHandler) SendMessageHandler(c *gin.Context) {
	// Get sender ID from token
	senderIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Convert senderID to int (might be float64 from JWT claims)
	var senderID int
	switch v := senderIDValue.(type) {
	case int:
		senderID = v
	case float64:
		senderID = int(v)
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID type"})
		return
	}

	// Parse request body
	var req domain.MessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	// Send message
	message, err := h.ChatUsecase.SendMessage(
		context.Background(),
		senderID,
		req.ReceiverID,
		req.Content,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Message sent successfully",
		"data":    message,
	})
}

// GetConversationMessagesHandler handles retrieving messages between two users
func (h *ChatHandler) GetConversationMessagesHandler(c *gin.Context) {
	// Get user ID from token
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Convert userID to int (might be float64 from JWT claims)
	var userID int
	switch v := userIDValue.(type) {
	case int:
		userID = v
	case float64:
		userID = int(v)
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID type"})
		return
	}

	// Get other user's ID from URL parameter
	otherUserIDStr := c.Param("user_id")
	otherUserID, err := strconv.Atoi(otherUserIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	// Get messages
	messages, err := h.ChatUsecase.GetConversationMessages(
		context.Background(),
		userID,
		otherUserID,
		limit,
		offset,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": messages,
	})
}

// GetUserConversationsHandler handles retrieving all conversations for a user
func (h *ChatHandler) GetUserConversationsHandler(c *gin.Context) {
	// Get user ID from token
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Convert userID to int (might be float64 from JWT claims)
	var userID int
	switch v := userIDValue.(type) {
	case int:
		userID = v
	case float64:
		userID = int(v)
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID type"})
		return
	}

	// Get conversations
	conversations, err := h.ChatUsecase.GetUserConversations(
		context.Background(),
		userID,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": conversations,
	})
}

package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go-auth-app/internal/delivery"
	"go-auth-app/internal/domain"
	"go-auth-app/internal/routes"
	"go-auth-app/internal/usecase"
	"go-auth-app/pkg"
)

// setupChatTestRouter creates a test router specifically for chat tests
func setupChatTestRouter() (*gin.Engine, *MockUserRepository, *MockChatRepository) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// Create mocks
	mockUserRepo := new(MockUserRepository)
	mockChatRepo := new(MockChatRepository)

	// Create usecases
	authUsecase := usecase.NewAuthUsecase(mockUserRepo)
	chatUsecase := usecase.NewChatUsecase(mockChatRepo, mockUserRepo, nil)

	// Create handlers
	authHandler := delivery.NewAuthHandler(authUsecase)
	chatHandler := delivery.NewChatHandler(chatUsecase)

	// Setup routes
	routes.SetupRoutes(router, authHandler, chatHandler, nil)

	return router, mockUserRepo, mockChatRepo
}

// TestSendMessage tests the chat/messages endpoint
func TestSendMessage(t *testing.T) {
	router, mockUserRepo, mockChatRepo := setupChatTestRouter()

	// Create JWT token for authentication
	token, _ := pkg.GenerateJWT(1, "test@example.com")

	// Mock user repository for receiver validation
	mockUserRepo.On("GetByID", mock.Anything, 2).Return(&domain.User{
		ID:    2,
		Name:  "Receiver User",
		Email: "receiver@example.com",
	}, nil)

	// Mock chat repository for message saving
	mockChatRepo.On("SaveMessage", mock.Anything, mock.AnythingOfType("*domain.Message")).Run(func(args mock.Arguments) {
		// Simulate setting an ID and creation time
		msg := args.Get(1).(*domain.Message)
		msg.ID = 1
		msg.CreatedAt = time.Now()
	}).Return(nil)

	// Mock conversation creation/retrieval
	mockChatRepo.On("GetOrCreateConversation", mock.Anything, 1, 2).Return(&domain.Conversation{
		ID:      1,
		User1ID: 1,
		User2ID: 2,
	}, nil)

	// Mock conversation update
	mockChatRepo.On("UpdateConversation", mock.Anything, 1, "Hello, receiver!").Return(nil)

	// Create message request
	messageReq := map[string]interface{}{
		"receiver_id": 2,
		"content":     "Hello, receiver!",
	}

	// Convert to JSON
	jsonData, _ := json.Marshal(messageReq)

	// Create request
	req, _ := http.NewRequest("POST", "/chat/messages", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify the response contains the expected structure
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check for message and data fields
	assert.Contains(t, response, "message")
	assert.Contains(t, response, "data")
	assert.Equal(t, "Message sent successfully", response["message"])

	mockUserRepo.AssertExpectations(t)
	mockChatRepo.AssertExpectations(t)
}

// TestSendMessageUnauthorized tests sending a message without authorization
func TestSendMessageUnauthorized(t *testing.T) {
	router, _, _ := setupChatTestRouter()

	// Create message request
	messageReq := map[string]interface{}{
		"receiver_id": 2,
		"content":     "Hello, receiver!",
	}

	// Convert to JSON
	jsonData, _ := json.Marshal(messageReq)

	// Create request without authentication token
	req, _ := http.NewRequest("POST", "/chat/messages", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestSendMessageInvalidRequest tests sending a message with invalid request data
func TestSendMessageInvalidRequest(t *testing.T) {
	router, _, _ := setupChatTestRouter()

	// Create JWT token for authentication
	token, _ := pkg.GenerateJWT(1, "test@example.com")

	// Create invalid message request (missing required fields)
	messageReq := map[string]interface{}{
		// Missing receiver_id
		"content": "Hello, receiver!",
	}

	// Convert to JSON
	jsonData, _ := json.Marshal(messageReq)

	// Create request
	req, _ := http.NewRequest("POST", "/chat/messages", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestGetConversationMessages tests retrieving messages between two users
func TestGetConversationMessages(t *testing.T) {
	router, mockUserRepo, mockChatRepo := setupChatTestRouter()

	// Create JWT token for authentication
	token, _ := pkg.GenerateJWT(1, "test@example.com")

	// Mock user repository for user validation
	mockUserRepo.On("GetByID", mock.Anything, 2).Return(&domain.User{
		ID:    2,
		Name:  "Other User",
		Email: "other@example.com",
	}, nil)

	// Mock conversation messages retrieval
	mockMessages := []*domain.Message{
		{
			ID:         1,
			SenderID:   1,
			ReceiverID: 2,
			Content:    "Hello from user 1",
			CreatedAt:  time.Now().Add(-time.Hour),
		},
		{
			ID:         2,
			SenderID:   2,
			ReceiverID: 1,
			Content:    "Hello from user 2",
			CreatedAt:  time.Now().Add(-30 * time.Minute),
		},
	}
	mockChatRepo.On("GetMessagesByConversation", mock.Anything, 1, 2, 20, 0).Return(mockMessages, nil)

	// Create request
	req, _ := http.NewRequest("GET", "/chat/messages/2?limit=20&offset=0", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify the response contains the expected structure
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check for data field
	assert.Contains(t, response, "data")

	// Verify data is an array with 2 messages
	data, ok := response["data"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(data))

	// Verify all mocks were called as expected
	mockUserRepo.AssertExpectations(t)
	mockChatRepo.AssertExpectations(t)
}

// TestGetUserConversations tests retrieving all conversations for a user
func TestGetUserConversations(t *testing.T) {
	router, _, mockChatRepo := setupChatTestRouter()

	// Create JWT token for authentication
	token, _ := pkg.GenerateJWT(1, "test@example.com")

	// Mock conversations retrieval
	mockConversations := []*domain.Conversation{
		{
			ID:          1,
			User1ID:     1,
			User2ID:     2,
			LastMessage: "Hello from user 2",
			UpdatedAt:   time.Now().Add(-30 * time.Minute),
		},
		{
			ID:          2,
			User1ID:     1,
			User2ID:     3,
			LastMessage: "How are you doing?",
			UpdatedAt:   time.Now().Add(-2 * time.Hour),
		},
	}
	mockChatRepo.On("GetConversationsByUserID", mock.Anything, 1).Return(mockConversations, nil)

	// Create request
	req, _ := http.NewRequest("GET", "/chat/conversations", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify the response contains the expected structure
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check for data field
	assert.Contains(t, response, "data")

	// Verify data is an array with 2 conversations
	data, ok := response["data"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(data))

	// Verify all mocks were called as expected
	mockChatRepo.AssertExpectations(t)
}

// TestSendMessageReceiverNotFound tests sending a message to a non-existent receiver
func TestSendMessageReceiverNotFound(t *testing.T) {
	router, mockUserRepo, _ := setupChatTestRouter()

	// Create JWT token for authentication
	token, _ := pkg.GenerateJWT(1, "test@example.com")

	// Mock user repository to return nil (user not found)
	mockUserRepo.On("GetByID", mock.Anything, 999).Return(nil, nil)

	// Create message request with non-existent receiver
	messageReq := map[string]interface{}{
		"receiver_id": 999,
		"content":     "Hello, non-existent user!",
	}

	// Convert to JSON
	jsonData, _ := json.Marshal(messageReq)

	// Create request
	req, _ := http.NewRequest("POST", "/chat/messages", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Verify mock was called
	mockUserRepo.AssertExpectations(t)
}

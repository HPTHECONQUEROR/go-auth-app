package tests

import (
	"net/http/httptest"
	"strings"
	"testing"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"go-auth-app/internal/delivery"
	"go-auth-app/internal/routes"
	"go-auth-app/internal/usecase"
)

// setupWebSocketTestRouter creates a test router specifically for WebSocket tests
func setupWebSocketTestRouter() (*gin.Engine, *MockUserRepository, *MockChatRepository, *httptest.Server) {
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
	wsHandler := delivery.NewWebSocketHandler(chatUsecase, nil)

	// Setup routes
	routes.SetupRoutes(router, authHandler, chatHandler, wsHandler)

	// Create test server
	server := httptest.NewServer(router)

	return router, mockUserRepo, mockChatRepo, server
}

// TestWebSocketAuthRequired tests that WebSocket connections require authentication
func TestWebSocketAuthRequired(t *testing.T) {
	_, _, _, server := setupWebSocketTestRouter()
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/chat/ws"
	_, _, err := websocket.DefaultDialer.Dial(wsURL, nil)

	// Connection should fail due to authentication failure
	assert.Error(t, err)
}


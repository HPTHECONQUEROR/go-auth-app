package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"go-auth-app/internal/delivery"
	"go-auth-app/internal/domain"
	"go-auth-app/internal/routes"
	"go-auth-app/internal/usecase"
	"go-auth-app/pkg"
)

// Mock repositories
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if user, ok := args.Get(0).(*domain.User); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int) (*domain.User, error) {
	args := m.Called(ctx, id)
	if user, ok := args.Get(0).(*domain.User); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1)
}

// Mock Chat Repository
type MockChatRepository struct {
	mock.Mock
}

func (m *MockChatRepository) SaveMessage(ctx context.Context, message *domain.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockChatRepository) GetMessagesByConversation(ctx context.Context, user1ID, user2ID int, limit, offset int) ([]*domain.Message, error) {
	args := m.Called(ctx, user1ID, user2ID, limit, offset)
	if messages, ok := args.Get(0).([]*domain.Message); ok {
		return messages, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockChatRepository) GetOrCreateConversation(ctx context.Context, user1ID, user2ID int) (*domain.Conversation, error) {
	args := m.Called(ctx, user1ID, user2ID)
	if conv, ok := args.Get(0).(*domain.Conversation); ok {
		return conv, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockChatRepository) GetConversationsByUserID(ctx context.Context, userID int) ([]*domain.Conversation, error) {
	args := m.Called(ctx, userID)
	if convs, ok := args.Get(0).([]*domain.Conversation); ok {
		return convs, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockChatRepository) UpdateConversation(ctx context.Context, conversationID int, lastMessage string) error {
	args := m.Called(ctx, conversationID, lastMessage)
	return args.Error(0)
}

// Mock NATS Service
type MockNATSService struct {
	mock.Mock
}

func (m *MockNATSService) PublishChatMessage(message *domain.Message) error {
	args := m.Called(message)
	return args.Error(0)
}

// Setup Test Router
func setupTestRouter() (*gin.Engine, *MockUserRepository, *MockChatRepository, *MockNATSService) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// Create mocks
	mockUserRepo := new(MockUserRepository)
	mockChatRepo := new(MockChatRepository)
	mockNATSService := new(MockNATSService)

	// Create usecases
	authUsecase := usecase.NewAuthUsecase(mockUserRepo)
	chatUsecase := usecase.NewChatUsecase(mockChatRepo, mockUserRepo, nil)

	// Create handlers
	authHandler := delivery.NewAuthHandler(authUsecase)
	chatHandler := delivery.NewChatHandler(chatUsecase)

	// Setup routes
	routes.SetupRoutes(router, authHandler, chatHandler, nil)

	return router, mockUserRepo, mockChatRepo, mockNATSService
}

// Signup Test
func TestSignup(t *testing.T) {
	router, mockUserRepo, _, _ := setupTestRouter()

	// Prepare test user
	user := map[string]string{
		"name":     "Test User",
		"email":    "test@example.com",
		"password": "securePassword123",
	}

	// Mock user repo expectations
	mockUserRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, nil)
	mockUserRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)

	// Convert user to JSON
	jsonUser, _ := json.Marshal(user)

	// Create request
	req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer(jsonUser))
	req.Header.Set("Content-Type", "application/json")

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusCreated, w.Code)
	mockUserRepo.AssertExpectations(t)
}

// Login Test
func TestLogin(t *testing.T) {
	router, mockUserRepo, _, _ := setupTestRouter()

	// Prepare login credentials
	loginCreds := map[string]string{
		"email":    "test@example.com",
		"password": "securePassword123",
	}

	// Create mock user with hashed password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("securePassword123"), bcrypt.DefaultCost)
	mockUser := &domain.User{
		ID:       1,
		Name:     "Test User",
		Email:    "test@example.com",
		Password: string(hashedPassword),
	}

	// Mock user repo expectations
	mockUserRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(mockUser, nil)

	// Convert login creds to JSON
	jsonCreds, _ := json.Marshal(loginCreds)

	// Create request
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonCreds))
	req.Header.Set("Content-Type", "application/json")

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	mockUserRepo.AssertExpectations(t)
}

// Protected Route Test
func TestProtectedRoute(t *testing.T) {
	router, _, _, _ := setupTestRouter()

	// Create a mock JWT token
	token, _ := pkg.GenerateJWT(1, "test@example.com")

	// Create request with JWT token
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
}


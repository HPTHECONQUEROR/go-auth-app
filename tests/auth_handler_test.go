package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"go-auth-app/internal/delivery"
	"go-auth-app/internal/domain"
	"go-auth-app/internal/repository"
	"go-auth-app/internal/usecase"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// Mock repository
type MockUserRepo struct {
	mock.Mock
}

// Ensure MockUserRepo implements repository.UserRepository
var _ repository.UserRepository = (*MockUserRepo)(nil)

// Implement Create method
func (m *MockUserRepo) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// Implement GetByEmail method
func (m *MockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if user, ok := args.Get(0).(*domain.User); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1)
}

// Hash password helper function
func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// Setup Test Environment
func setupTestRouter() (*gin.Engine, *MockUserRepo) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	mockRepo := new(MockUserRepo)
	authUsecase := usecase.NewAuthUsecase(mockRepo)
	authHandler := delivery.NewAuthHandler(authUsecase)

	router.POST("/signup", authHandler.SignupHandler)
	router.POST("/login", authHandler.LoginHandler)

	return router, mockRepo
}

// Test Signup
func TestSignUp(t *testing.T) {
	router, mockRepo := setupTestRouter()

	mockUser := &domain.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	// Ensure GetByEmail is mocked before calling Signup
	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, nil)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)

	reqBody, _ := json.Marshal(mockUser)
	req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockRepo.AssertExpectations(t)
}

// Test Login
func TestLogin(t *testing.T) {
	router, mockRepo := setupTestRouter()

	// Hash password before storing it in mock
	hashedPassword, _ := hashPassword("password1234567")

	mockUser := &domain.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: hashedPassword,
	}

	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(mockUser, nil)

	loginData := map[string]string{
		"email":    "test@example.com",
		"password": "password1234567",
	}

	reqBody, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}
package delivery

import (
	"context"
	"fmt"
	"go-auth-app/internal/domain"
	"go-auth-app/internal/usecase"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	AuthUsecase *usecase.AuthUsecase
}

func NewAuthHandler(authUsecase *usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{AuthUsecase: authUsecase}
}

//signup-handler
func (h *AuthHandler) SignupHandler(c *gin.Context) {
	var user domain.User

	if err := c.ShouldBindJSON(&user); err != nil {
		// Get the error message
		errorMsg := err.Error()

		// Custom error messages based on the error content
		if strings.Contains(errorMsg, "Name") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "The name field must not be empty but the given Name Is not in valid format.", "details": errorMsg})
		} else if strings.Contains(errorMsg, "Email") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "The Email must be in correct format check the email you have provided", "details": errorMsg})
		} else if strings.Contains(errorMsg, "Password") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password must have atleast 10 characters", "details": errorMsg})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": errorMsg})
		}
		return
	}

	fmt.Printf("Received User: %+v\n", user)

	if err := h.AuthUsecase.Signup(context.Background(), &user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

//login-handler
func (h *AuthHandler) LoginHandler(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"` // ← fixed missing quote
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		errorMsg := err.Error()

		switch {
		case strings.Contains(errorMsg, "Email"):
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Email is required and must be in valid format.",
				"details": errorMsg,
			})
		case strings.Contains(errorMsg, "Password"):
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Password is required and must be provided.",
				"details": errorMsg,
			})
		default:
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid input. Please check your credentials.",
				"details": errorMsg,
			})
		}
		return
	}

	token, err := h.AuthUsecase.Login(context.Background(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Authentication failed. Please check your email and password.",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}


func (h *AuthHandler) ProtectedHandler(c *gin.Context) {
	userID := c.GetInt("id")
	email := c.GetString("email")
	c.JSON(http.StatusOK, gin.H{
		"message": "Protected Data",
		"user_id": userID,
		"email":   email,
	})
}

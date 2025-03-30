package delivery

import (
	"fmt"
	"context"
	"go-auth-app/internal/domain"
	"go-auth-app/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	AuthUsecase *usecase.AuthUsecase
}

func NewAuthHandler(authUsecase *usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{AuthUsecase: authUsecase}
}


func (h *AuthHandler) SignupHandler(c *gin.Context) {
    var user domain.User

    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
        return
    }

    fmt.Printf("Received User: %+v\n", user)

    if err := h.AuthUsecase.Signup(context.Background(), &user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

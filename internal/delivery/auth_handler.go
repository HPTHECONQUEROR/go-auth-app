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

//signup-handler
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


//login-handler

func (h *AuthHandler) LoginHandler(c *gin.Context){
    var req struct{
        Email string `json:"email"`
        Password string `json:"password`
    }

    if err := c.ShouldBindJSON(&req); err != nil{
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})  
    
        return
    }

    token, err := h.AuthUsecase.Login(context.Background(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})

}
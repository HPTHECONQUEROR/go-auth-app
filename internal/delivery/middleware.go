package delivery

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go-auth-app/pkg"
)

// ErrorResponse function defines the standard error response structure

type ErrorResponse struct{
	Message string `json:"message"`
}


//ErrorHandlerMiddleware Fucntion handles error globally

func ErrorHandlerMiddleware() gin.HandlerFunc{
	return func (c *gin.Context)  {
		c.Next()

		if len(c.Errors) > 0{
			err := c.Errors.Last()
			c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
			c.Abort()
		}
		
	}
}



//AuthMiddleware Function validates the JWT token and extracts user information

func AuthMiddleware() gin.HandlerFunc{
	return func (c *gin.Context)  {
		authHeader := c.GetHeader("Authorization")
		if authHeader == ""{
			c.JSON(http.StatusUnauthorized,gin.H{
				"error":"Authorization header is missing",
			})
			c.Abort()
			return
		}

		// Extract token (format: "Bearer <token>")
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}
		
		// Validate token
		claims, err := pkg.ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims["user_id"])
		c.Set("email", claims["email"])

		c.Next()
		
	}

}
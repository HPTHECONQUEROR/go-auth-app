package pkg

import (
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var secretKey = []byte(os.Getenv("JWT_SECRET"))

func GenerateJWT(userID int,email string)(string,error){
	claims := jwt.MapClaims{
		"user_id":userID,
		"email":email,
		"exp":time.Now().Add(time.Hour*24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256,claims)
	return token.SignedString(secretKey)
}
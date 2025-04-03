package domain

import (
	"errors"
	// "fmt"
	"regexp"
	"time"
)

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name" binding:"required"`
	Email     string    `json:"email" binding:"required,email"`
	Password  string    `json:"password" binding:"required,min=10"`
	CreatedAt time.Time `json:"created_at"`
}

func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

func (u *User) Validate() error {
	if u.Name == "" {
		return errors.New("Name is required")
	}
	if u.Email == "" {
		return errors.New("Email is required")
	}
	if u.Password == "" {
		return errors.New("Password is required")
	}
	if len(u.Password) <= 10 {
		return errors.New("Password cannot be less than 10 characters!")
	}
	if !isValidEmail(u.Email) {
		return errors.New("The email is not valid!")
	}
	return nil
}

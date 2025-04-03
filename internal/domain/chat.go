package domain

import (
	"errors"
	"time"
)

// Message represents a chat message between users
type Message struct {
	ID         int       `json:"id"`
	SenderID   int       `json:"sender_id"`
	ReceiverID int       `json:"receiver_id"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}

// Conversation represents a chat conversation between two users
type Conversation struct {
	ID          int       `json:"id"`
	User1ID     int       `json:"user1_id"`
	User2ID     int       `json:"user2_id"`
	LastMessage string    `json:"last_message"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MessageRequest is used for receiving message data from clients
type MessageRequest struct {
	ReceiverID int    `json:"receiver_id" binding:"required"`
	Content    string `json:"content" binding:"required"`
}

// ValidateMessage validates the message content
func ValidateMessage(content string) error {
	if content == "" {
		return errors.New("message content cannot be empty")
	}
	return nil
}

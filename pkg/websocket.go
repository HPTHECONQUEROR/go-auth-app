package pkg

import (
	"encoding/json"
	"github.com/gorilla/websocket"
)

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// Client represents a connected WebSocket client
type Client struct {
	// The websocket connection
	Conn *websocket.Conn
	// Buffered channel of outbound messages
	Send chan WebSocketMessage
	// UserID from authentication
	ID int
}

// NewClient creates a new WebSocket client
func NewClient(conn *websocket.Conn, userID int) *Client {
	return &Client{
		Conn: conn,
		Send: make(chan WebSocketMessage, 256),
		ID:   userID,
	}
}

// MessageTypes constants
const (
	// Message types
	TypeChat          = "chat"
	TypeChatConfirmed = "chat_confirmed"
	TypeError         = "error"
)

// ChatMessage represents a chat message sent over WebSocket
type ChatMessage struct {
	SenderID   int    `json:"sender_id"`
	ReceiverID int    `json:"receiver_id"`
	Content    string `json:"content"`
}

// ErrorMessage represents an error message
type ErrorMessage struct {
	Message string `json:"message"`
}
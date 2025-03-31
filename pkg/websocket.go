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
	TypeChat           = "chat"
	TypeChatConfirmed  = "chat_confirmed"
	TypeTyping         = "typing"
	TypeRead           = "read"
	TypeError          = "error"
	TypeUserConnected  = "user_connected"
	TypeUserDisconnected = "user_disconnected"
)

// ChatMessage represents a chat message sent over WebSocket
type ChatMessage struct {
	SenderID   int    `json:"sender_id"`
	ReceiverID int    `json:"receiver_id"`
	Content    string `json:"content"`
	MessageID  int    `json:"message_id,omitempty"`
	Timestamp  int64  `json:"timestamp,omitempty"`
}

// TypingNotification represents a typing notification
type TypingNotification struct {
	SenderID   int  `json:"sender_id"`
	ReceiverID int  `json:"receiver_id"`
	IsTyping   bool `json:"is_typing"`
}

// ReadReceipt represents a read receipt
type ReadReceipt struct {
	MessageID int `json:"message_id"`
	ReaderID  int `json:"reader_id"`
}

// ErrorMessage represents an error message
type ErrorMessage struct {
	Message string `json:"message"`
}

// CreateChatMessage creates a WebSocketMessage of type chat
func CreateChatMessage(msg ChatMessage) WebSocketMessage {
	data, _ := json.Marshal(msg)
	return WebSocketMessage{
		Type: TypeChat,
		Data: data,
	}
}

// CreateTypingNotification creates a WebSocketMessage of type typing
func CreateTypingNotification(notification TypingNotification) WebSocketMessage {
	data, _ := json.Marshal(notification)
	return WebSocketMessage{
		Type: TypeTyping,
		Data: data,
	}
}

// CreateReadReceipt creates a WebSocketMessage of type read
func CreateReadReceipt(receipt ReadReceipt) WebSocketMessage {
	data, _ := json.Marshal(receipt)
	return WebSocketMessage{
		Type: TypeRead,
		Data: data,
	}
}

// CreateErrorMessage creates a WebSocketMessage of type error
func CreateErrorMessage(message string) WebSocketMessage {
	data, _ := json.Marshal(ErrorMessage{Message: message})
	return WebSocketMessage{
		Type: TypeError,
		Data: data,
	}
}
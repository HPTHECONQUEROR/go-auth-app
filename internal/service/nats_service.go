package service

import (
	"encoding/json"
	"fmt"
	"go-auth-app/internal/domain"
	"go-auth-app/pkg"
	"log"
	// "strconv"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
)

// NATSService provides chat-specific NATS functionality
type NATSService struct {
	Client        *pkg.NatsClient
	Subscriptions map[string]*nats.Subscription
}

// NATSMessagePayload is the structure of messages published over NATS
type NATSMessagePayload struct {
	ID         int    `json:"id"`
	SenderID   int    `json:"sender_id"`
	ReceiverID int    `json:"receiver_id"`
	Content    string `json:"content"`
	Timestamp  int64  `json:"timestamp"`
}

// NewNATSService creates a new NATS service
func NewNATSService(client *pkg.NatsClient) *NATSService {
	return &NATSService{
		Client:        client,
		Subscriptions: make(map[string]*nats.Subscription),
	}
}

// GetPrivateChatSubject returns the canonical subject name for a private chat between two users
func (s *NATSService) GetPrivateChatSubject(user1ID, user2ID int) string {
	// Ensure we always use the same ordering of user IDs for consistent subject naming
	if user1ID > user2ID {
		user1ID, user2ID = user2ID, user1ID
	}
	return fmt.Sprintf("chat.private.%d.%d", user1ID, user2ID)
}

// PublishChatMessage publishes a chat message to NATS
func (s *NATSService) PublishChatMessage(message *domain.Message) error {
	if !s.Client.IsConnected() {
		return fmt.Errorf("not connected to NATS")
	}

	// Create payload
	payload := NATSMessagePayload{
		ID:         message.ID,
		SenderID:   message.SenderID,
		ReceiverID: message.ReceiverID,
		Content:    message.Content,
		Timestamp:  message.CreatedAt.Unix(),
	}

	// Marshal to JSON
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal message payload: %v", err)
	}

	// Get subject
	subject := s.GetPrivateChatSubject(message.SenderID, message.ReceiverID)

	// Publish message
	err = s.Client.Publish(subject, data)
	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	log.Printf("Published message to subject: %s", subject)
	return nil
}

// SubscribeToUserMessages subscribes to all messages for a specific user
func (s *NATSService) SubscribeToUserMessages(userID int, callback func(senderID, receiverID int, content string)) error {
	if !s.Client.IsConnected() {
		return fmt.Errorf("not connected to NATS")
	}

	// Pattern for receiving messages where the user is either sender or receiver
	// This subscribes to both patterns:
	// 1. chat.private.<userID>.* - User is the smaller ID
	// 2. chat.private.*.<userID> - User is the larger ID
	subject := fmt.Sprintf("chat.private.*.%d", userID)
	subKey := fmt.Sprintf("user_%d_larger", userID)

	// Handle received messages - case where user is the larger ID
	messageCb := func(msg *nats.Msg) {
		s.handleChatMessage(msg, userID, callback)
	}

	// Subscribe to pattern where user is the larger ID
	sub, err := s.Client.Subscribe(subject, messageCb)
	if err != nil {
		return fmt.Errorf("failed to subscribe to subject %s: %v", subject, err)
	}
	s.Subscriptions[subKey] = sub

	// Also subscribe to pattern where user is the smaller ID
	subject = fmt.Sprintf("chat.private.%d.*", userID)
	subKey = fmt.Sprintf("user_%d_smaller", userID)
	
	sub, err = s.Client.Subscribe(subject, messageCb)
	if err != nil {
		return fmt.Errorf("failed to subscribe to subject %s: %v", subject, err)
	}
	s.Subscriptions[subKey] = sub

	log.Printf("Subscribed to all messages for user %d", userID)
	return nil
}

// handleChatMessage processes NATS messages for chat
func (s *NATSService) handleChatMessage(msg *nats.Msg, userID int, callback func(senderID, receiverID int, content string)) {
	// Parse subject to identify participants
	parts := strings.Split(msg.Subject, ".")
	if len(parts) != 4 {
		log.Printf("Invalid subject format: %s", msg.Subject)
		return
	}

	// Parse message payload - we only need the payload data, not subject user IDs
	var payload NATSMessagePayload
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		log.Printf("Failed to unmarshal message payload: %v", err)
		return
	}

	// Verify that the message is actually intended for this user
	if payload.SenderID != userID && payload.ReceiverID != userID {
		log.Printf("Received message not intended for user %d", userID)
		return
	}

	// Call the callback with the message data
	callback(payload.SenderID, payload.ReceiverID, payload.Content)
}

// UnsubscribeAll unsubscribes from all subscriptions
func (s *NATSService) UnsubscribeAll() {
	for key, sub := range s.Subscriptions {
		if err := s.Client.Unsubscribe(sub); err != nil {
			log.Printf("Failed to unsubscribe from %s: %v", key, err)
		}
	}
	s.Subscriptions = make(map[string]*nats.Subscription)
}

// Mock data for testing
func (s *NATSService) MockChatData() {
	// Generate some mock data for testing
	log.Println("Publishing mock chat data to NATS...")
	
	mockMessages := []struct {
		senderID   int
		receiverID int
		content    string
	}{
		{1, 2, "Hello from user 1 to user 2"},
		{2, 1, "Hello from user 2 to user 1"},
		{1, 2, "How are you doing?"},
		{2, 1, "I'm doing great, thanks for asking!"},
	}
	
	for _, msg := range mockMessages {
		mockMessage := &domain.Message{
			ID:         0, // Will be ignored for mock
			SenderID:   msg.senderID,
			ReceiverID: msg.receiverID,
			Content:    msg.content,
			CreatedAt:  time.Now(),
		}
		
		if err := s.PublishChatMessage(mockMessage); err != nil {
			log.Printf("Failed to publish mock message: %v", err)
		} else {
			log.Printf("Published mock message: %s", msg.content)
		}
	}
}
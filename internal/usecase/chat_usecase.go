package usecase

import (
	"context"
	"errors"
	"go-auth-app/internal/domain"
	"go-auth-app/internal/repository"
	"go-auth-app/internal/service"
	"log"
)

// ChatUsecase handles business logic for chat operations
type ChatUsecase struct {
	ChatRepo    repository.ChatRepository
	UserRepo    repository.UserRepository
	NatsService *service.NATSService
}

// NewChatUsecase creates a new instance of ChatUsecase
func NewChatUsecase(
	chatRepo repository.ChatRepository,
	userRepo repository.UserRepository,
	natsService *service.NATSService,
) *ChatUsecase {
	return &ChatUsecase{
		ChatRepo:    chatRepo,
		UserRepo:    userRepo,
		NatsService: natsService,
	}
}

// SendMessage sends a message from one user to another
func (uc *ChatUsecase) SendMessage(ctx context.Context, senderID int, receiverID int, content string) (*domain.Message, error) {
	// Validate message content
	if err := domain.ValidateMessage(content); err != nil {
		return nil, err
	}

	// Check if receiver exists
	receiver, err := uc.UserRepo.GetByID(ctx, receiverID)
	if err != nil {
		return nil, errors.New("receiver not found")
	}

	if receiver == nil {
		return nil, errors.New("receiver not found")
	}

	// Create message object
	message := &domain.Message{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    content,
	}

	// Save message to database
	err = uc.ChatRepo.SaveMessage(ctx, message)
	if err != nil {
		return nil, err
	}

	// Update conversation
	conversation, err := uc.ChatRepo.GetOrCreateConversation(ctx, senderID, receiverID)
	if err != nil {
		return nil, err
	}

	err = uc.ChatRepo.UpdateConversation(ctx, conversation.ID, content)
	if err != nil {
		return nil, err
	}

	// Publish message to NATS
	if uc.NatsService != nil {
		err = uc.NatsService.PublishChatMessage(message)
		if err != nil {
			// Log error but don't fail the operation
			log.Printf("Failed to publish message to NATS: %v", err)
		}
	}

	return message, nil
}

// GetConversationMessages retrieves messages between two users
func (uc *ChatUsecase) GetConversationMessages(ctx context.Context, user1ID int, user2ID int, limit int, offset int) ([]*domain.Message, error) {
	// Check if user2 exists
	user2, err := uc.UserRepo.GetByID(ctx, user2ID)
	if err != nil || user2 == nil {
		return nil, errors.New("user not found")
	}

	// Set default pagination values
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	// Get messages
	messages, err := uc.ChatRepo.GetMessagesByConversation(ctx, user1ID, user2ID, limit, offset)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

// GetUserConversations retrieves all conversations for a user
func (uc *ChatUsecase) GetUserConversations(ctx context.Context, userID int) ([]*domain.Conversation, error) {
	conversations, err := uc.ChatRepo.GetConversationsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return conversations, nil
}

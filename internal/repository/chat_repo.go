package repository

import (
	"context"
	"go-auth-app/db"
	"go-auth-app/internal/domain"
	"time"
)

// ChatRepository defines the interface for chat-related operations
type ChatRepository interface {
	// Message methods
	SaveMessage(ctx context.Context, message *domain.Message) error
	GetMessagesByConversation(ctx context.Context, user1ID, user2ID int, limit, offset int) ([]*domain.Message, error)
	GetMessageByID(ctx context.Context, messageID int) (*domain.Message, error)
	MarkMessageAsRead(ctx context.Context, messageID int) error
	
	// Conversation methods
	GetOrCreateConversation(ctx context.Context, user1ID, user2ID int) (*domain.Conversation, error)
	GetConversationsByUserID(ctx context.Context, userID int) ([]*domain.Conversation, error)
	UpdateConversation(ctx context.Context, conversationID int, lastMessage string) error
}

// chatRepo implements ChatRepository
type chatRepo struct{}

// NewChatRepository creates a new instance of chatRepo
func NewChatRepository() ChatRepository {
	return &chatRepo{}
}

// SaveMessage stores a new message in the database
func (r *chatRepo) SaveMessage(ctx context.Context, message *domain.Message) error {
	query := `
		INSERT INTO messages (sender_id, receiver_id, content, created_at, read)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	
	now := time.Now()
	err := db.DB.QueryRow(
		ctx,
		query,
		message.SenderID,
		message.ReceiverID,
		message.Content,
		now,
		false,
	).Scan(&message.ID)
	
	if err != nil {
		return err
	}
	
	message.CreatedAt = now
	message.Read = false
	
	// Update the conversation with the latest message
	_, err = r.GetOrCreateConversation(ctx, message.SenderID, message.ReceiverID)
	if err != nil {
		return err
	}
	
	return nil
}

// GetMessagesByConversation retrieves messages between two users with pagination
func (r *chatRepo) GetMessagesByConversation(ctx context.Context, user1ID, user2ID int, limit, offset int) ([]*domain.Message, error) {
	query := `
		SELECT id, sender_id, receiver_id, content, created_at, read
		FROM messages
		WHERE (sender_id = $1 AND receiver_id = $2) OR (sender_id = $2 AND receiver_id = $1)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`
	
	rows, err := db.DB.Query(ctx, query, user1ID, user2ID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	messages := []*domain.Message{}
	for rows.Next() {
		msg := &domain.Message{}
		err := rows.Scan(
			&msg.ID,
			&msg.SenderID,
			&msg.ReceiverID,
			&msg.Content,
			&msg.CreatedAt,
			&msg.Read,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	return messages, nil
}

// GetMessageByID retrieves a specific message by its ID
func (r *chatRepo) GetMessageByID(ctx context.Context, messageID int) (*domain.Message, error) {
	query := `
		SELECT id, sender_id, receiver_id, content, created_at, read
		FROM messages
		WHERE id = $1
	`
	
	msg := &domain.Message{}
	err := db.DB.QueryRow(ctx, query, messageID).Scan(
		&msg.ID,
		&msg.SenderID,
		&msg.ReceiverID,
		&msg.Content,
		&msg.CreatedAt,
		&msg.Read,
	)
	
	if err != nil {
		return nil, err
	}
	
	return msg, nil
}

// MarkMessageAsRead marks a message as read
func (r *chatRepo) MarkMessageAsRead(ctx context.Context, messageID int) error {
	query := `
		UPDATE messages
		SET read = true
		WHERE id = $1
	`
	
	_, err := db.DB.Exec(ctx, query, messageID)
	return err
}

// GetOrCreateConversation gets an existing conversation or creates a new one
func (r *chatRepo) GetOrCreateConversation(ctx context.Context, user1ID, user2ID int) (*domain.Conversation, error) {
	// First try to get existing conversation
	query := `
		SELECT id, user1_id, user2_id, last_message, updated_at, created_at
		FROM conversations
		WHERE (user1_id = $1 AND user2_id = $2) OR (user1_id = $2 AND user2_id = $1)
	`
	
	conversation := &domain.Conversation{}
	err := db.DB.QueryRow(ctx, query, user1ID, user2ID).Scan(
		&conversation.ID,
		&conversation.User1ID,
		&conversation.User2ID,
		&conversation.LastMessage,
		&conversation.UpdatedAt,
		&conversation.CreatedAt,
	)
	
	if err != nil {
		// If not found, create a new conversation
		createQuery := `
			INSERT INTO conversations (user1_id, user2_id, last_message, updated_at, created_at)
			VALUES ($1, $2, $3, $4, $4)
			RETURNING id, created_at
		`
		
		now := time.Now()
		err = db.DB.QueryRow(
			ctx,
			createQuery,
			user1ID,
			user2ID,
			"",
			now,
		).Scan(&conversation.ID, &conversation.CreatedAt)
		
		if err != nil {
			return nil, err
		}
		
		conversation.User1ID = user1ID
		conversation.User2ID = user2ID
		conversation.LastMessage = ""
		conversation.UpdatedAt = now
	}
	
	return conversation, nil
}

// GetConversationsByUserID retrieves all conversations for a user
func (r *chatRepo) GetConversationsByUserID(ctx context.Context, userID int) ([]*domain.Conversation, error) {
	query := `
		SELECT id, user1_id, user2_id, last_message, updated_at, created_at
		FROM conversations
		WHERE user1_id = $1 OR user2_id = $1
		ORDER BY updated_at DESC
	`
	
	rows, err := db.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	conversations := []*domain.Conversation{}
	for rows.Next() {
		conv := &domain.Conversation{}
		err := rows.Scan(
			&conv.ID,
			&conv.User1ID,
			&conv.User2ID,
			&conv.LastMessage,
			&conv.UpdatedAt,
			&conv.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		conversations = append(conversations, conv)
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	return conversations, nil
}

// UpdateConversation updates the conversation with the latest message
func (r *chatRepo) UpdateConversation(ctx context.Context, conversationID int, lastMessage string) error {
	query := `
		UPDATE conversations
		SET last_message = $1, updated_at = $2
		WHERE id = $3
	`
	
	_, err := db.DB.Exec(ctx, query, lastMessage, time.Now(), conversationID)
	return err
}
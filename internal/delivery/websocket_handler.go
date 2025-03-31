package delivery

import (
	"context"
	"encoding/json"
	"go-auth-app/internal/domain"
	"go-auth-app/internal/usecase"
	"go-auth-app/pkg"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocketHandler handles WebSocket connections for real-time chat
type WebSocketHandler struct {
	ChatUsecase *usecase.ChatUsecase
	// Track active connections
	clients    map[int]*pkg.Client
	clientsMux sync.RWMutex
	// WebSocket upgrader
	upgrader websocket.Upgrader
}

// NewWebSocketHandler creates a new instance of WebSocketHandler
func NewWebSocketHandler(chatUsecase *usecase.ChatUsecase) *WebSocketHandler {
	return &WebSocketHandler{
		ChatUsecase: chatUsecase,
		clients:     make(map[int]*pkg.Client),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			// Allow all origins for development. In production, this should be restricted.
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// HandleWebSocket upgrades the HTTP connection to WebSocket
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// Get user ID from token
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Convert userID to int (might be float64 from JWT claims)
	var userID int
	switch v := userIDValue.(type) {
	case int:
		userID = v
	case float64:
		userID = int(v)
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID type"})
		return
	}

	// Upgrade connection to WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Error upgrading to WebSocket: %v", err)
		return
	}

	// Create client
	client := pkg.NewClient(conn, userID)

	// Register client
	h.registerClient(client)

	// Start client handlers
	go h.handleMessages(client)
	go h.handleClientConnection(client)
}

// registerClient adds a client to the clients map
func (h *WebSocketHandler) registerClient(client *pkg.Client) {
	h.clientsMux.Lock()
	defer h.clientsMux.Unlock()
	h.clients[client.ID] = client
	log.Printf("Client connected: %d", client.ID)
}

// unregisterClient removes a client from the clients map
func (h *WebSocketHandler) unregisterClient(client *pkg.Client) {
	h.clientsMux.Lock()
	defer h.clientsMux.Unlock()
	if _, ok := h.clients[client.ID]; ok {
		delete(h.clients, client.ID)
		client.Conn.Close()
		log.Printf("Client disconnected: %d", client.ID)
	}
}

// getClient gets a client by ID
func (h *WebSocketHandler) getClient(id int) *pkg.Client {
	h.clientsMux.RLock()
	defer h.clientsMux.RUnlock()
	return h.clients[id]
}

// handleClientConnection handles a client's WebSocket connection
func (h *WebSocketHandler) handleClientConnection(client *pkg.Client) {
	defer func() {
		h.unregisterClient(client)
	}()

	// Keep the connection alive with ping/pong
	client.Conn.SetPingHandler(func(appData string) error {
		err := client.Conn.WriteControl(websocket.PongMessage, []byte{}, time.Now().Add(10*time.Second))
		if err == websocket.ErrCloseSent {
			return nil
		}
		return err
	})

	// Main read loop
	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error reading message: %v", err)
			}
			break
		}

		// Parse message
		var wsMessage pkg.WebSocketMessage
		if err := json.Unmarshal(message, &wsMessage); err != nil {
			log.Printf("Error parsing message: %v", err)
			continue
		}

		// Process message based on type
		switch wsMessage.Type {
		case "chat":
			h.handleChatMessage(client, wsMessage.Data)
		case "typing":
			h.handleTypingNotification(client, wsMessage.Data)
		case "read":
			h.handleReadReceipt(client, wsMessage.Data)
		}
	}
}

// handleMessages processes messages from the message channel
func (h *WebSocketHandler) handleMessages(client *pkg.Client) {
	for message := range client.Send {
		err := client.Conn.WriteJSON(message)
		if err != nil {
			log.Printf("Error sending message to client %d: %v", client.ID, err)
			client.Conn.Close()
			break
		}
	}
}

// handleChatMessage processes a chat message
func (h *WebSocketHandler) handleChatMessage(client *pkg.Client, data json.RawMessage) {
	var msgReq domain.MessageRequest
	if err := json.Unmarshal(data, &msgReq); err != nil {
		log.Printf("Error parsing chat message: %v", err)
		return
	}

	// Store message in database
	message, err := h.ChatUsecase.SendMessage(
		context.Background(),
		client.ID,
		msgReq.ReceiverID,
		msgReq.Content,
	)
	if err != nil {
		log.Printf("Error saving message: %v", err)
		// Send error message back to sender
		errorData, _ := json.Marshal(map[string]string{"message": err.Error()})
		errorMsg := pkg.WebSocketMessage{
			Type: "error",
			Data: errorData,
		}
		client.Send <- errorMsg
		return
	}

	// Send message to recipient if they're online
	if recipient := h.getClient(msgReq.ReceiverID); recipient != nil {
		messageData, _ := json.Marshal(message)
		wsMessage := pkg.WebSocketMessage{
			Type: "chat",
			Data: messageData,
		}
		recipient.Send <- wsMessage
	}

	// Send confirmation to sender
	messageData, _ := json.Marshal(message)
	confirmMsg := pkg.WebSocketMessage{
		Type: "chat_confirmed",
		Data: messageData,
	}
	client.Send <- confirmMsg
}

// handleTypingNotification processes a typing notification
func (h *WebSocketHandler) handleTypingNotification(client *pkg.Client, data json.RawMessage) {
	var typingData struct {
		ReceiverID int  `json:"receiver_id"`
		IsTyping   bool `json:"is_typing"`
	}
	if err := json.Unmarshal(data, &typingData); err != nil {
		log.Printf("Error parsing typing notification: %v", err)
		return
	}

	// Forward typing notification to recipient if they're online
	if recipient := h.getClient(typingData.ReceiverID); recipient != nil {
		typingResponse := struct {
			SenderID int  `json:"sender_id"`
			IsTyping bool `json:"is_typing"`
		}{
			SenderID: client.ID,
			IsTyping: typingData.IsTyping,
		}
		data, err := json.Marshal(typingResponse)
		if err != nil {
			log.Printf("Error marshaling typing notification: %v", err)
			return
		}
		
		wsMessage := pkg.WebSocketMessage{
			Type: "typing",
			Data: data,
		}
		recipient.Send <- wsMessage
	}
}

// handleReadReceipt processes a read receipt
func (h *WebSocketHandler) handleReadReceipt(client *pkg.Client, data json.RawMessage) {
	var readData struct {
		MessageID int `json:"message_id"`
	}
	if err := json.Unmarshal(data, &readData); err != nil {
		log.Printf("Error parsing read receipt: %v", err)
		return
	}

	// Mark message as read in database
	err := h.ChatUsecase.MarkMessageAsRead(
		context.Background(),
		readData.MessageID,
		client.ID,
	)
	if err != nil {
		log.Printf("Error marking message as read: %v", err)
		return
	}

	// Get the message to identify the sender
	message, err := h.ChatUsecase.ChatRepo.GetMessageByID(context.Background(), readData.MessageID)
	if err != nil {
		log.Printf("Error getting message for read receipt: %v", err)
		return
	}

	// Send read receipt to sender if they're online
	if sender := h.getClient(message.SenderID); sender != nil {
		readResponse := struct {
			MessageID int `json:"message_id"`
			ReaderID  int `json:"reader_id"`
		}{
			MessageID: readData.MessageID,
			ReaderID:  client.ID,
		}
		data, err := json.Marshal(readResponse)
		if err != nil {
			log.Printf("Error marshaling read receipt: %v", err)
			return
		}
		
		wsMessage := pkg.WebSocketMessage{
			Type: "read",
			Data: data,
		}
		sender.Send <- wsMessage
	}
}
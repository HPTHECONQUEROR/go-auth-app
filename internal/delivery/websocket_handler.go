package delivery

import (
	"context"
	"encoding/json"
	"go-auth-app/internal/domain"
	"go-auth-app/internal/service"
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
	NatsService *service.NATSService
	// Track active connections
	clients    map[int]*pkg.Client
	clientsMux sync.RWMutex
	// WebSocket upgrader
	upgrader websocket.Upgrader
}

// NewWebSocketHandler creates a new instance of WebSocketHandler
func NewWebSocketHandler(chatUsecase *usecase.ChatUsecase, natsService *service.NATSService) *WebSocketHandler {
	return &WebSocketHandler{
		ChatUsecase: chatUsecase,
		NatsService: natsService,
		clients:     make(map[int]*pkg.Client),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			// Allow all origins for development
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

	// Subscribe to NATS for this user
	if h.NatsService != nil {
		err = h.NatsService.SubscribeToUserMessages(userID, func(senderID, receiverID int, content string) {
			// Only forward messages if they're intended for this user
			if receiverID == userID {
				// Convert to a domain message
				message := &domain.Message{
					SenderID:   senderID,
					ReceiverID: receiverID,
					Content:    content,
					CreatedAt:  time.Now(),
				}

				// Marshal message to JSON for WebSocket transport
				messageData, err := json.Marshal(message)
				if err != nil {
					log.Printf("Error marshaling NATS message: %v", err)
					return
				}

				// Create WebSocket message
				wsMessage := pkg.WebSocketMessage{
					Type: "chat",
					Data: messageData,
				}

				// Send to client
				client.Send <- wsMessage
				log.Printf("Forwarded NATS message to user %d", userID)
			}
		})

		if err != nil {
			log.Printf("Error subscribing to NATS for user %d: %v", userID, err)
		}
	}

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
		if wsMessage.Type == "chat" {
			h.handleChatMessage(client, wsMessage.Data)
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

	// Store message in database and publish to NATS
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

	// Send confirmation to sender
	messageData, _ := json.Marshal(message)
	confirmMsg := pkg.WebSocketMessage{
		Type: "chat_confirmed",
		Data: messageData,
	}
	client.Send <- confirmMsg

	// Note: We don't need to manually send to recipient here anymore
	// The message will be delivered via NATS subscription
}

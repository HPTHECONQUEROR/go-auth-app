package pkg

import (
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

// NatsClient manages the NATS connection and provides methods for pub/sub
type NatsClient struct {
	Conn      *nats.Conn
	URL       string
	Reconnect bool
}

// NewNatsClient creates a new NATS client
func NewNatsClient(url string, reconnect bool) (*NatsClient, error) {
	client := &NatsClient{
		URL:       url,
		Reconnect: reconnect,
	}

	err := client.Connect()
	if err != nil {
		return nil, err
	}

	return client, nil
}

// Connect establishes a connection to the NATS server
func (n *NatsClient) Connect() error {
	// NATS connection options
	opts := []nats.Option{
		nats.Name("Go Auth Chat Service"),
		nats.ReconnectWait(5 * time.Second),
		nats.MaxReconnects(-1), // Unlimited reconnects
	}

	// Add reconnection handlers
	if n.Reconnect {
		opts = append(opts, nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Printf("NATS disconnected: %v", err)
		}))

		opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("NATS reconnected to %s", nc.ConnectedUrl())
		}))

		opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
			log.Printf("NATS connection closed")
		}))
	}

	// Connect to NATS
	conn, err := nats.Connect(n.URL, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %v", err)
	}

	n.Conn = conn
	log.Printf("Connected to NATS at %s", n.URL)
	return nil
}

// Close closes the NATS connection
func (n *NatsClient) Close() {
	if n.Conn != nil {
		n.Conn.Close()
		log.Println("NATS connection closed")
	}
}

// Publish publishes a message to the specified subject
func (n *NatsClient) Publish(subject string, data []byte) error {
	if n.Conn == nil {
		return fmt.Errorf("not connected to NATS")
	}

	err := n.Conn.Publish(subject, data)
	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	return nil
}

// Subscribe subscribes to the specified subject and calls the provided callback for each message
func (n *NatsClient) Subscribe(subject string, callback func(*nats.Msg)) (*nats.Subscription, error) {
	if n.Conn == nil {
		return nil, fmt.Errorf("not connected to NATS")
	}

	sub, err := n.Conn.Subscribe(subject, callback)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to subject %s: %v", subject, err)
	}

	log.Printf("Subscribed to subject: %s", subject)
	return sub, nil
}

// QueueSubscribe subscribes to the specified subject as part of a queue group
func (n *NatsClient) QueueSubscribe(subject, queue string, callback func(*nats.Msg)) (*nats.Subscription, error) {
	if n.Conn == nil {
		return nil, fmt.Errorf("not connected to NATS")
	}

	sub, err := n.Conn.QueueSubscribe(subject, queue, callback)
	if err != nil {
		return nil, fmt.Errorf("failed to queue subscribe to subject %s: %v", subject, err)
	}

	log.Printf("Queue subscribed to subject: %s with queue: %s", subject, queue)
	return sub, nil
}

// Unsubscribe unsubscribes from a subscription
func (n *NatsClient) Unsubscribe(sub *nats.Subscription) error {
	if sub == nil {
		return fmt.Errorf("subscription is nil")
	}

	err := sub.Unsubscribe()
	if err != nil {
		return fmt.Errorf("failed to unsubscribe: %v", err)
	}

	log.Printf("Unsubscribed from subject: %s", sub.Subject)
	return nil
}

// IsConnected returns true if connected to NATS
func (n *NatsClient) IsConnected() bool {
	return n.Conn != nil && n.Conn.IsConnected()
}

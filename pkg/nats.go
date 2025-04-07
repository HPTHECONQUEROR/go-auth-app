package pkg

import (
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

type NatsClient struct {
	Conn      *nats.Conn
	URL       string
	Reconnect bool
}

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

func (n *NatsClient) Connect() error {
	opts := []nats.Option{
		nats.Name("Go Auth NATS Client"),
		nats.ReconnectWait(5 * time.Second),
		nats.MaxReconnects(-1),
	}

	if n.Reconnect {
		opts = append(opts, 
			nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
				log.Printf("NATS disconnected: %v", err)
			}),
			nats.ReconnectHandler(func(nc *nats.Conn) {
				log.Printf("NATS reconnected to %s", nc.ConnectedUrl())
			}),
		)
	}

	conn, err := nats.Connect(n.URL, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %v", err)
	}

	n.Conn = conn
	log.Printf("Connected to NATS at %s", n.URL)
	return nil
}

func (n *NatsClient) Close() {
	if n.Conn != nil {
		n.Conn.Close()
		log.Println("NATS connection closed")
	}
}

func (n *NatsClient) Publish(subject string, data []byte) error {
	if n.Conn == nil {
		return fmt.Errorf("not connected to NATS")
	}

	return n.Conn.Publish(subject, data)
}

// IsConnected checks if the client is currently connected to NATS
func (n *NatsClient) IsConnected() bool {
	return n.Conn != nil && n.Conn.IsConnected()
}

// Subscribe wraps the NATS subscription process
func (n *NatsClient) Subscribe(subject string, cb nats.MsgHandler) (*nats.Subscription, error) {
	if n.Conn == nil {
		return nil, fmt.Errorf("not connected to NATS")
	}
	
	return n.Conn.Subscribe(subject, cb)
}

// Unsubscribe handles unsubscribing from a NATS subscription
func (n *NatsClient) Unsubscribe(sub *nats.Subscription) error {
	if sub == nil {
		return fmt.Errorf("subscription is nil")
	}
	
	return sub.Unsubscribe()
}
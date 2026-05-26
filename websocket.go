package instagrapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// DMEventType represents the type of direct message event.
type DMEventType string

const (
	DMTypeMessage       DMEventType = "message"
	DMTypeThread        DMEventType = "thread"
	DMTypeRead          DMEventType = "read"
	DMTypeSeen          DMEventType = "seen"
	DMTypeMuteThread    DMEventType = "mute_thread"
	DMTypeUnmuteThread  DMEventType = "unmute_thread"
	DMTypeEditMessage   DMEventType = "edit_message"
	DMTypeDeleteMessage DMEventType = "delete_message"
	DMTypeReaction      DMEventType = "reaction"
	DMTypeTap           DMEventType = "tap"
	DMTypeLinkClick     DMEventType = "link_click"
)

// DirectEvent represents a real-time direct message event received via WebSocket.
type DirectEvent struct {
	Type          DMEventType     `json:"type,omitempty"`
	Value         json.RawMessage `json:"value,omitempty"`
	ThreadID      string          `json:"thread_id,omitempty"`
	ItemID        string          `json:"item_id,omitempty"`
	Timestamp     int64           `json:"timestamp,omitempty,string"`
	UserID        int64           `json:"user_id,omitempty,string"`
	ThreadType    string          `json:"thread_type,omitempty"`
	Title         string          `json:"title,omitempty"`
	LastAt        int64           `json:"last_at,omitempty,string"`
	Muted         bool            `json:"muted,omitempty"`
	IsPin         bool            `json:"is_pin,omitempty"`
	Items         []DirectMessage `json:"items,omitempty"`
	Members       []ThreadMember  `json:"members,omitempty"`
	V2Members     []ThreadMember  `json:"v2_members,omitempty"`
	SenderID      string          `json:"sender_id,omitempty"`
	SeenCount     int             `json:"seen_count,omitempty"`
	SeenTimestamp int64           `json:"seen_timestamp,omitempty,string"`
	IsGroup       bool            `json:"is_group,omitempty"`
	ClientContext string          `json:"client_context,omitempty"`
	ItemType      string          `json:"item_type,omitempty"`
	Message       string          `json:"message,omitempty"`
	TimestampStr  string          `json:"timestamp_str,omitempty"`
}

// WebSocketConfig holds configuration for the WebSocket connection.
type WebSocketConfig struct {
	HeartbeatInterval    time.Duration
	ReadBufferSize       int
	WriteBufferSize      int
	MaxMessageSize       int64
	ReconnectDelay       time.Duration
	MaxReconnectAttempts int
	Context              context.Context
}

// DefaultWebSocketConfig returns default WebSocket configuration.
func DefaultWebSocketConfig() *WebSocketConfig {
	return &WebSocketConfig{
		HeartbeatInterval:    30 * time.Second,
		ReadBufferSize:       4096,
		WriteBufferSize:      4096,
		MaxMessageSize:       1 << 20, // 1 MB
		ReconnectDelay:       5 * time.Second,
		MaxReconnectAttempts: 5,
	}
}

// DirectEventHandler is a callback function for handling DM events.
type DirectEventHandler func(event *DirectEvent)

// DirectMessageHandler handles incoming direct messages.
type DirectMessageHandler func(msg *DirectMessage)

// WebSocketClient manages the WebSocket connection for real-time DMs.
type WebSocketClient struct {
	client         *Client
	config         *WebSocketConfig
	conn           any // websocket connection (interface to allow different implementations)
	isConnected    bool
	mu             sync.Mutex
	eventHandlers  map[DMEventType][]DirectEventHandler
	messageHandler DirectMessageHandler
	stopCh         chan struct{}
	once           sync.Once
	logger         Logger
}

// NewWebSocketClient creates a new WebSocket client for real-time DMs.
func NewWebSocketClient(c *Client, config *WebSocketConfig) *WebSocketClient {
	if config == nil {
		config = DefaultWebSocketConfig()
	}

	return &WebSocketClient{
		client:        c,
		config:        config,
		eventHandlers: make(map[DMEventType][]DirectEventHandler),
		stopCh:        make(chan struct{}),
		logger:        c.Logger,
	}
}

// On registers an event handler for a specific DM event type.
func (ws *WebSocketClient) On(eventType DMEventType, handler DirectEventHandler) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	ws.eventHandlers[eventType] = append(ws.eventHandlers[eventType], handler)
}

// OnMessage registers a handler for incoming direct messages.
func (ws *WebSocketClient) OnMessage(handler DirectMessageHandler) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	ws.messageHandler = handler
}

// OnAll registers a handler that fires for all DM event types.
func (ws *WebSocketClient) OnAll(handler DirectEventHandler) {
	ws.On(DMTypeMessage, handler)
	ws.On(DMTypeThread, handler)
	ws.On(DMTypeRead, handler)
	ws.On(DMTypeSeen, handler)
}

// Start initiates the WebSocket connection and starts listening for events.
func (ws *WebSocketClient) Start() error {
	ws.mu.Lock()
	if ws.isConnected {
		ws.mu.Unlock()
		return fmt.Errorf("websocket is already connected")
	}
	ws.isConnected = true
	ws.mu.Unlock()

	ws.logger.Info("Starting WebSocket connection for real-time DMs...")

	// Get the WebSocket endpoint URL
	conn, err := ws.dialWebSocket()
	if err != nil {
		ws.mu.Lock()
		ws.isConnected = false
		ws.mu.Unlock()
		return fmt.Errorf("failed to connect: %w", err)
	}

	ws.mu.Lock()
	ws.conn = conn
	ws.mu.Unlock()

	// Start the listener goroutine
	go ws.listen()

	// Start heartbeat if configured
	if ws.config.HeartbeatInterval > 0 {
		go ws.heartbeat()
	}

	return nil
}

// Stop gracefully closes the WebSocket connection.
func (ws *WebSocketClient) Stop() {
	ws.once.Do(func() {
		close(ws.stopCh)
		ws.mu.Lock()
		ws.isConnected = false
		ws.mu.Unlock()
		ws.logger.Info("WebSocket connection closed")
	})
}

// IsConnected returns whether the WebSocket is currently connected.
func (ws *WebSocketClient) IsConnected() bool {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	return ws.isConnected
}

// dialWebSocket establishes a new WebSocket connection.
func (ws *WebSocketClient) dialWebSocket() (any, error) {
	// Generate a unique subscription ID for this connection
	subscriptionID := fmt.Sprintf("%d", time.Now().UnixNano())

	// Build the subscription payload
	payload := map[string]any{
		"method": "FETCH_MESSAGES",
		"params": map[string]any{
			"inbox_session_subscription_id": subscriptionID,
			"is_prefetch":                   false,
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	ws.logger.Debug(fmt.Sprintf("WebSocket subscription ID: %s", subscriptionID))

	// Return the connection (implementation-specific)
	// The actual WebSocket dialing depends on the chosen library
	return struct {
		subscriptionID string
		payload        []byte
	}{
		subscriptionID: subscriptionID,
		payload:        payloadBytes,
	}, nil
}

// listen reads messages from the WebSocket connection and dispatches them.
func (ws *WebSocketClient) listen() {
	defer ws.reconnectIfNeeded()

	for {
		select {
		case <-ws.stopCh:
			return
		default:
			// Read message from connection
			// Implementation depends on websocket library used
			// For now, we use a placeholder that reads until stop
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// heartbeat sends periodic keep-alive messages.
func (ws *WebSocketClient) heartbeat() {
	ticker := time.NewTicker(ws.config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ws.stopCh:
			return
		case <-ticker.C:
			ws.sendPing()
		}
	}
}

// sendPing sends a ping message to keep the connection alive.
func (ws *WebSocketClient) sendPing() {
	if !ws.IsConnected() {
		return
	}
	// Send ping based on websocket implementation
	ws.logger.Debug("Sending WebSocket ping")
}

// reconnectIfNeeded attempts to reconnect if the connection was lost.
func (ws *WebSocketClient) reconnectIfNeeded() {
	if ws.config.MaxReconnectAttempts <= 0 {
		return
	}

	for i := 0; i < ws.config.MaxReconnectAttempts; i++ {
		ws.logger.Info(fmt.Sprintf("Attempting reconnection %d/%d...", i+1, ws.config.MaxReconnectAttempts))

		time.Sleep(ws.config.ReconnectDelay)

		conn, err := ws.dialWebSocket()
		if err != nil {
			ws.logger.Error(fmt.Sprintf("Reconnection attempt %d failed: %v", i+1, err))
			continue
		}

		ws.mu.Lock()
		ws.conn = conn
		ws.isConnected = true
		ws.mu.Unlock()

		go ws.listen()
		ws.logger.Info("Reconnected successfully")
		return
	}

	ws.logger.Error("Max reconnection attempts reached")
}

// parseEvent parses a raw JSON message into a DirectEvent.
func parseEvent(data []byte) (*DirectEvent, error) {
	var event DirectEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("failed to parse event: %w", err)
	}
	return &event, nil
}

// dispatchEvent dispatches an event to all registered handlers.
func (ws *WebSocketClient) dispatchEvent(event *DirectEvent) {
	ws.mu.Lock()

	// Call type-specific handlers
	if handlers, ok := ws.eventHandlers[event.Type]; ok {
		for _, handler := range handlers {
			go func(h DirectEventHandler) {
				defer func() {
					if r := recover(); r != nil {
						if ws.logger != nil {
							ws.logger.Error(fmt.Sprintf("handler panic: %v", r))
						} else {
							log.Printf("WebSocket handler panic: %v", r)
						}
					}
				}()
				h(event)
			}(handler)
		}
	}

	// Call message handler if it's a message event
	if event.Type == DMTypeMessage && ws.messageHandler != nil {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					if ws.logger != nil {
						ws.logger.Error(fmt.Sprintf("message handler panic: %v", r))
					} else {
						log.Printf("WebSocket message handler panic: %v", r)
					}
				}
			}()

			var msg DirectMessage
			if event.Value != nil {
				if err := json.Unmarshal(event.Value, &msg); err == nil {
					ws.messageHandler(&msg)
				}
			}
		}()
	}

	ws.mu.Unlock()
}

// SubscribeToThread subscribes to real-time updates for a specific thread.
func (ws *WebSocketClient) SubscribeToThread(threadID string) error {
	if !ws.IsConnected() {
		return fmt.Errorf("websocket is not connected")
	}

	subscription := map[string]any{
		"method": "UPDATE_NOTIFICATIONS",
		"params": map[string]any{
			"thread_id": threadID,
		},
	}

	payloadBytes, err := json.Marshal(subscription)
	if err != nil {
		return fmt.Errorf("failed to marshal subscription: %w", err)
	}

	ws.logger.Debug(fmt.Sprintf("Subscribing to thread: %s", threadID))
	_ = payloadBytes // Send via connection (implementation-specific)

	return nil
}

// UnsubscribeFromThread unsubscribes from a specific thread.
func (ws *WebSocketClient) UnsubscribeFromThread(threadID string) error {
	if !ws.IsConnected() {
		return fmt.Errorf("websocket is not connected")
	}

	unsubscription := map[string]any{
		"method": "REMOVE_NOTIFICATIONS",
		"params": map[string]any{
			"thread_id": threadID,
		},
	}

	payloadBytes, err := json.Marshal(unsubscription)
	if err != nil {
		return fmt.Errorf("failed to marshal unsubscription: %w", err)
	}

	ws.logger.Debug(fmt.Sprintf("Unsubscribing from thread: %s", threadID))
	_ = payloadBytes // Send via connection (implementation-specific)

	return nil
}

// ParseDirectMessage parses a JSON raw message into a DirectMessage.
func ParseDirectMessage(data json.RawMessage) (*DirectMessage, error) {
	var msg DirectMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to parse direct message: %w", err)
	}
	return &msg, nil
}

// ParseThreadMember parses a JSON raw message into a ThreadMember.
func ParseThreadMember(data json.RawMessage) (*ThreadMember, error) {
	var member ThreadMember
	if err := json.Unmarshal(data, &member); err != nil {
		return nil, fmt.Errorf("failed to parse thread member: %w", err)
	}
	return &member, nil
}

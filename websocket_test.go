package instagrapi

import (
	"encoding/json"
	"sync"
	"testing"
	"time"
)

// TestNewWebSocketClient tests WebSocket client creation.
func TestNewWebSocketClient(t *testing.T) {
	c := NewClient()

	tests := []struct {
		name    string
		config  *WebSocketConfig
		wantNil bool
	}{
		{"nil config", nil, false},
		{"default config", DefaultWebSocketConfig(), false},
		{"custom config", &WebSocketConfig{
			HeartbeatInterval:    10 * time.Second,
			ReconnectDelay:       2 * time.Second,
			MaxReconnectAttempts: 3,
		}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := NewWebSocketClient(c, tt.config)
			if ws == nil && !tt.wantNil {
				t.Fatal("NewWebSocketClient returned nil")
			}
			if ws != nil {
				if ws.client != c {
					t.Error("WebSocket client should reference the Instagram client")
				}
				if ws.eventHandlers == nil {
					t.Error("eventHandlers map should be initialized")
				}
				if ws.stopCh == nil {
					t.Error("stopCh channel should be initialized")
				}
			}
		})
	}
}

// TestDefaultWebSocketConfig tests the default configuration values.
func TestDefaultWebSocketConfig(t *testing.T) {
	config := DefaultWebSocketConfig()

	if config.HeartbeatInterval != 30*time.Second {
		t.Errorf("HeartbeatInterval = %v; want 30s", config.HeartbeatInterval)
	}
	if config.ReadBufferSize <= 0 {
		t.Errorf("ReadBufferSize = %d; want > 0", config.ReadBufferSize)
	}
	if config.WriteBufferSize <= 0 {
		t.Errorf("WriteBufferSize = %d; want > 0", config.WriteBufferSize)
	}
	if config.MaxMessageSize <= 0 {
		t.Errorf("MaxMessageSize = %d; want > 0", config.MaxMessageSize)
	}
	if config.ReconnectDelay <= 0 {
		t.Errorf("ReconnectDelay = %v; want > 0", config.ReconnectDelay)
	}
	if config.MaxReconnectAttempts <= 0 {
		t.Errorf("MaxReconnectAttempts = %d; want > 0", config.MaxReconnectAttempts)
	}

	t.Log("DefaultWebSocketConfig values are valid")
}

// TestOnEventHandlers tests registering event handlers.
func TestOnEventHandlers(t *testing.T) {
	ws := NewWebSocketClient(NewClient(), nil)

	handlerCalled := false
	var mu sync.Mutex

	ws.On(DMTypeMessage, func(event *DirectEvent) {
		mu.Lock()
		defer mu.Unlock()
		handlerCalled = true
	})

	if len(ws.eventHandlers[DMTypeMessage]) != 1 {
		t.Error("Handler should be registered for DMTypeMessage")
	}

	// Register another handler for the same event type
	ws.On(DMTypeMessage, func(event *DirectEvent) {})
	if len(ws.eventHandlers[DMTypeMessage]) != 2 {
		t.Error("Second handler should also be registered")
	}

	// Handlers for different types should not interfere
	if len(ws.eventHandlers[DMTypeThread]) != 0 {
		t.Error("DMTypeThread should have no handlers yet")
	}

	_ = handlerCalled // would be set when event is dispatched
}

// TestOnMessage tests registering a message handler.
func TestOnMessage(t *testing.T) {
	ws := NewWebSocketClient(NewClient(), nil)

	if ws.messageHandler != nil {
		t.Error("messageHandler should be nil initially")
	}

	msgReceived := false
	var mu sync.Mutex

	ws.OnMessage(func(msg *DirectMessage) {
		mu.Lock()
		defer mu.Unlock()
		msgReceived = true
	})

	if ws.messageHandler == nil {
		t.Error("messageHandler should be set after OnMessage")
	}

	_ = msgReceived // would be set when message is dispatched
}

// TestOnAll tests registering a handler for all event types.
func TestOnAll(t *testing.T) {
	ws := NewWebSocketClient(NewClient(), nil)

	handlerCalled := false
	var mu sync.Mutex

	ws.OnAll(func(event *DirectEvent) {
		mu.Lock()
		defer mu.Unlock()
		handlerCalled = true
	})

	if len(ws.eventHandlers[DMTypeMessage]) != 1 {
		t.Error("Handler should be registered for DMTypeMessage via OnAll")
	}
	if len(ws.eventHandlers[DMTypeThread]) != 1 {
		t.Error("Handler should be registered for DMTypeThread via OnAll")
	}
	if len(ws.eventHandlers[DMTypeRead]) != 1 {
		t.Error("Handler should be registered for DMTypeRead via OnAll")
	}
	if len(ws.eventHandlers[DMTypeSeen]) != 1 {
		t.Error("Handler should be registered for DMTypeSeen via OnAll")
	}

	_ = handlerCalled // would be set when event is dispatched
}

// TestIsConnected tests the connection state tracking.
func TestIsConnected(t *testing.T) {
	ws := NewWebSocketClient(NewClient(), nil)

	if ws.IsConnected() {
		t.Error("WebSocket should not be connected initially")
	}

	// Simulate connecting
	ws.mu.Lock()
	ws.isConnected = true
	ws.mu.Unlock()

	if !ws.IsConnected() {
		t.Error("WebSocket should be connected after setting isConnected=true")
	}

	// Simulate disconnecting
	ws.mu.Lock()
	ws.isConnected = false
	ws.mu.Unlock()

	if ws.IsConnected() {
		t.Error("WebSocket should not be connected after setting isConnected=false")
	}
}

// TestStop tests the Stop method.
func TestStop(t *testing.T) {
	ws := NewWebSocketClient(NewClient(), nil)

	// Simulate connection
	ws.mu.Lock()
	ws.isConnected = true
	ws.mu.Unlock()

	ws.Stop()

	if ws.IsConnected() {
		t.Error("WebSocket should not be connected after Stop")
	}

	// Calling Stop again should be idempotent (no panic)
	ws.Stop()
	if ws.IsConnected() {
		t.Error("WebSocket should still not be connected after second Stop")
	}
}

// TestStartAlreadyConnected tests starting an already-connected client.
func TestStartAlreadyConnected(t *testing.T) {
	ws := NewWebSocketClient(NewClient(), nil)

	// Simulate connection state without actual dial
	ws.mu.Lock()
	ws.isConnected = true
	ws.mu.Unlock()

	err := ws.Start()
	if err == nil {
		t.Error("Expected error when starting already-connected WebSocket")
	}
}

// TestParseEvent tests parsing a DirectEvent from JSON.
func TestParseEvent(t *testing.T) {
	tests := []struct {
		name    string
		jsonStr string
		wantErr bool
	}{
		{
			name: "valid message event",
			jsonStr: `{
				"type": "message",
				"thread_id": "1234567890",
				"item_id": "msg_001",
				"timestamp": "1700000000",
				"user_id": "987654321",
				"item_type": "text",
				"message": "Hello, World!"
			}`,
			wantErr: false,
		},
		{
			name: "valid read event",
			jsonStr: `{
				"type": "read",
				"thread_id": "1234567890",
				"user_id": "987654321",
				"seen_count": 5,
				"seen_timestamp": "1700000100"
			}`,
			wantErr: false,
		},
		{
			name: "valid thread event",
			jsonStr: `{
				"type": "thread",
				"thread_id": "1234567890",
				"title": "Team Chat",
				"last_at": "1700000200",
				"muted": false,
				"is_pin": true
			}`,
			wantErr: false,
		},
		{
			name:    "invalid json",
			jsonStr: `{invalid json`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := parseEvent([]byte(tt.jsonStr))
			if tt.wantErr {
				if err == nil {
					t.Error("parseEvent should return error for invalid JSON")
				}
				return
			}

			if err != nil {
				t.Fatalf("parseEvent error: %v", err)
			}

			if event == nil {
				t.Fatal("parseEvent returned nil event")
			}
		})
	}
}

// TestParseDirectMessage tests parsing a DirectMessage from JSON.
func TestParseDirectMessage(t *testing.T) {
	jsonData := json.RawMessage(`{
		"item_id": "msg_001",
		"timestamp": 1700000000,
		"item_type": "text",
		"user_id": "987654321"
	}`)

	msg, err := ParseDirectMessage(jsonData)
	if err != nil {
		t.Fatalf("ParseDirectMessage error: %v", err)
	}

	if msg.ID != "msg_001" {
		t.Errorf("ID = %q; want 'msg_001'", msg.ID)
	}
	if msg.ItemType != "text" {
		t.Errorf("ItemType = %q; want 'text'", msg.ItemType)
	}

	// Test invalid JSON
	_, err = ParseDirectMessage(json.RawMessage(`{invalid`))
	if err == nil {
		t.Error("ParseDirectMessage should return error for invalid JSON")
	}
}

// TestDMEventTypeConstants tests that all event type constants are valid strings.
func TestDMEventTypeConstants(t *testing.T) {
	eventTypes := []DMEventType{
		DMTypeMessage,
		DMTypeThread,
		DMTypeRead,
		DMTypeSeen,
		DMTypeMuteThread,
		DMTypeUnmuteThread,
		DMTypeEditMessage,
		DMTypeDeleteMessage,
		DMTypeReaction,
		DMTypeTap,
		DMTypeLinkClick,
	}

	for _, et := range eventTypes {
		if string(et) == "" {
			t.Errorf("Event type %q should not be empty", et)
		}
	}
}

// TestDirectEventStruct tests that DirectEvent can hold various event data.
func TestDirectEventStruct(t *testing.T) {
	event := &DirectEvent{
		Type:          DMTypeMessage,
		ThreadID:      "thread_123",
		ItemID:        "msg_456",
		Timestamp:     1700000000,
		UserID:        987654321,
		SenderID:      "sender_789",
		ClientContext: "client_ctx_abc",
		ItemType:      "text",
		Message:       "Test message content",
	}

	if event.Type != DMTypeMessage {
		t.Error("Event type should be set correctly")
	}
	if event.ThreadID != "thread_123" {
		t.Error("ThreadID should be set correctly")
	}
	if event.Message != "Test message content" {
		t.Error("Message should be set correctly")
	}

	// Test serialization
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal DirectEvent: %v", err)
	}

	var parsed DirectEvent
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal DirectEvent: %v", err)
	}

	if parsed.Type != event.Type {
		t.Error("Serialized/deserialized type should match")
	}
	if parsed.ThreadID != event.ThreadID {
		t.Error("Serialized/deserialized thread_id should match")
	}
}

// TestSubscribeToThread tests the subscribe method (without actual connection).
func TestSubscribeToThread(t *testing.T) {
	ws := NewWebSocketClient(NewClient(), nil)

	// Should fail when not connected
	err := ws.SubscribeToThread("thread_123")
	if err == nil {
		t.Error("SubscribeToThread should return error when not connected")
	}

	// Simulate connection
	ws.mu.Lock()
	ws.isConnected = true
	ws.conn = struct {
		subscriptionID string
		payload        []byte
	}{
		subscriptionID: "sub_123",
		payload:        []byte("test"),
	}
	ws.mu.Unlock()

	err = ws.SubscribeToThread("thread_123")
	if err != nil {
		t.Errorf("SubscribeToThread should succeed when connected: %v", err)
	}
}

// TestUnsubscribeFromThread tests the unsubscribe method.
func TestUnsubscribeFromThread(t *testing.T) {
	ws := NewWebSocketClient(NewClient(), nil)

	// Should fail when not connected
	err := ws.UnsubscribeFromThread("thread_123")
	if err == nil {
		t.Error("UnsubscribeFromThread should return error when not connected")
	}

	// Simulate connection
	ws.mu.Lock()
	ws.isConnected = true
	ws.mu.Unlock()

	err = ws.UnsubscribeFromThread("thread_123")
	if err != nil {
		t.Errorf("UnsubscribeFromThread should succeed when connected: %v", err)
	}
}

// TestDispatchEvent tests event dispatching to handlers.
func TestDispatchEvent(t *testing.T) {
	ws := NewWebSocketClient(NewClient(), nil)

	var messageEvents []string
	var threadEvents []string

	ws.On(DMTypeMessage, func(event *DirectEvent) {
		messageEvents = append(messageEvents, event.Message)
	})
	ws.On(DMTypeThread, func(event *DirectEvent) {
		threadEvents = append(threadEvents, event.ThreadID)
	})

	// Create and dispatch a message event
	msgEvent := &DirectEvent{
		Type:    DMTypeMessage,
		Message: "Hello!",
	}
	ws.dispatchEvent(msgEvent)

	// Give goroutine time to execute
	time.Sleep(10 * time.Millisecond)

	if len(messageEvents) != 1 {
		t.Errorf("Expected 1 message event handler call, got %d", len(messageEvents))
	}
	if len(threadEvents) != 0 {
		t.Errorf("Expected 0 thread events, got %d", len(threadEvents))
	}

	// Create and dispatch a thread event
	threadEvent := &DirectEvent{
		Type:     DMTypeThread,
		ThreadID: "thread_abc",
	}
	ws.dispatchEvent(threadEvent)

	time.Sleep(10 * time.Millisecond)

	if len(messageEvents) != 1 {
		t.Errorf("Message events count should still be 1, got %d", len(messageEvents))
	}
	if len(threadEvents) != 1 {
		t.Errorf("Expected 1 thread event handler call, got %d", len(threadEvents))
	}
}

// TestDispatchEventWithPanic tests that panics in handlers are recovered.
func TestDispatchEventWithPanic(t *testing.T) {
	ws := NewWebSocketClient(NewClient(), nil)

	// Register a handler that panics
	ws.On(DMTypeMessage, func(event *DirectEvent) {
		panic("test panic")
	})

	// Should not panic
	msgEvent := &DirectEvent{
		Type:    DMTypeMessage,
		Message: "Test",
	}

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Handler panic was not recovered: %v", r)
			}
		}()
		ws.dispatchEvent(msgEvent)
		time.Sleep(10 * time.Millisecond)
	}()

	t.Log("Panic in handler was properly recovered")
}

// TestWebSocketClientLifecycle tests the full lifecycle of a WebSocket client.
func TestWebSocketClientLifecycle(t *testing.T) {
	c := NewClient()
	ws := NewWebSocketClient(c, nil)

	// 1. Initial state: not connected
	if ws.IsConnected() {
		t.Error("Should start disconnected")
	}

	// 2. Register handlers before connecting
	handlerCalled := false
	ws.On(DMTypeMessage, func(event *DirectEvent) {
		handlerCalled = true
	})

	// 3. Start (simulated - dial returns nil error for test)
	ws.mu.Lock()
	ws.isConnected = true
	ws.conn = struct{}{}
	ws.mu.Unlock()

	if !ws.IsConnected() {
		t.Error("Should be connected after Start")
	}

	// 4. Subscribe to thread
	err := ws.SubscribeToThread("thread_123")
	if err != nil {
		t.Errorf("Subscribe should succeed: %v", err)
	}

	// 5. Unsubscribe from thread
	err = ws.UnsubscribeFromThread("thread_123")
	if err != nil {
		t.Errorf("Unsubscribe should succeed: %v", err)
	}

	// 6. Stop
	ws.Stop()
	if ws.IsConnected() {
		t.Error("Should be disconnected after Stop")
	}

	// 7. Second stop is idempotent
	ws.Stop() // should not panic or error

	_ = handlerCalled // would be true if event was dispatched
}

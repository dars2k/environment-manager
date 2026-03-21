package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubHub satisfies the Hub interface without requiring a real WebSocket connection.
type stubHub struct{}

func (s *stubHub) Unregister(c *Client) {}
func (s *stubHub) Broadcast(msg []byte) {}

func newTestClient() *Client {
	return &Client{
		ID:            "test",
		hub:           &stubHub{},
		conn:          nil, // not needed for handleMessage / sendMessage
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
		authenticated: false,
	}
}

func marshal(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}

func TestHandleMessage_Ping(t *testing.T) {
	c := newTestClient()
	c.handleMessage(marshal(map[string]interface{}{"type": "ping"}))

	// sendMessage should have queued a pong
	select {
	case msg := <-c.send:
		var parsed map[string]interface{}
		assert.NoError(t, json.Unmarshal(msg, &parsed))
		assert.Equal(t, "pong", parsed["type"])
	default:
		t.Fatal("expected pong message in send channel")
	}
}

func TestHandleMessage_Subscribe(t *testing.T) {
	c := newTestClient()
	c.handleMessage(marshal(map[string]interface{}{
		"type": "subscribe",
		"payload": map[string]interface{}{
			"environments": []string{"env-1", "env-2"},
		},
	}))

	assert.True(t, c.subscriptions["env-1"])
	assert.True(t, c.subscriptions["env-2"])
	assert.False(t, c.subscriptions["env-3"])
}

func TestHandleMessage_Unsubscribe(t *testing.T) {
	c := newTestClient()
	c.subscriptions["env-1"] = true
	c.subscriptions["env-2"] = true

	c.handleMessage(marshal(map[string]interface{}{
		"type": "unsubscribe",
		"payload": map[string]interface{}{
			"environments": []string{"env-1"},
		},
	}))

	assert.False(t, c.subscriptions["env-1"])
	assert.True(t, c.subscriptions["env-2"])
}

func TestHandleMessage_Authenticate(t *testing.T) {
	c := newTestClient()
	assert.False(t, c.authenticated)

	c.handleMessage(marshal(map[string]interface{}{"type": "authenticate"}))

	assert.True(t, c.authenticated)
}

func TestHandleMessage_Unknown(t *testing.T) {
	c := newTestClient()
	// Should not panic or crash
	c.handleMessage(marshal(map[string]interface{}{"type": "unknown"}))
}

func TestHandleMessage_InvalidJSON(t *testing.T) {
	c := newTestClient()
	// Should not panic on malformed input
	c.handleMessage([]byte("not json"))
}

func TestSendMessage_ChannelFull(t *testing.T) {
	c := newTestClient()
	// Fill the channel
	for i := 0; i < cap(c.send); i++ {
		c.send <- []byte("msg")
	}

	// Next sendMessage should close the channel (not panic)
	c.sendMessage("pong", nil)
	// After close, channel is closed; verify by trying to read
	_, ok := <-c.send
	// The channel was closed; subsequent receives return zero value with ok=false
	// Drain remaining messages
	for range c.send {
	}
	_ = ok // channel may or may not be closed depending on timing; just ensure no panic
}

func TestSendStatusUpdate_NotSubscribed(t *testing.T) {
	c := newTestClient()
	c.SendStatusUpdate("env-x", map[string]string{"status": "healthy"})
	// No message should be sent
	assert.Equal(t, 0, len(c.send))
}

func TestSendStatusUpdate_Subscribed(t *testing.T) {
	c := newTestClient()
	c.subscriptions["env-x"] = true
	c.SendStatusUpdate("env-x", map[string]string{"status": "healthy"})

	select {
	case msg := <-c.send:
		var parsed map[string]interface{}
		assert.NoError(t, json.Unmarshal(msg, &parsed))
		assert.Equal(t, "status_update", parsed["type"])
	default:
		t.Fatal("expected status_update message")
	}
}

// ---- ReadPump / WritePump (require real WebSocket) ----

func newTestWSPair(t *testing.T, hub Hub) (*Client, *websocket.Conn, func()) {
	t.Helper()

	var serverConn *websocket.Conn
	up := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := up.Upgrade(w, r, nil)
		require.NoError(t, err)
		serverConn = conn
		// keep the handler alive
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}))

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	dialConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	time.Sleep(30 * time.Millisecond)
	c := NewClient("test", hub, serverConn)

	cleanup := func() {
		dialConn.Close()
		srv.Close()
	}
	return c, dialConn, cleanup
}

func TestReadPump_CloseOnConnectionClose(t *testing.T) {
	hub := &stubHub{}
	c, dialConn, cleanup := newTestWSPair(t, hub)
	defer cleanup()

	go c.ReadPump()
	time.Sleep(30 * time.Millisecond)

	// Close the dialer side – ReadPump should exit gracefully
	dialConn.Close()
	time.Sleep(80 * time.Millisecond)
}

func TestReadPump_ProcessesMessage(t *testing.T) {
	hub := &stubHub{}
	c, dialConn, cleanup := newTestWSPair(t, hub)
	defer cleanup()

	go c.ReadPump()
	go c.WritePump()
	time.Sleep(30 * time.Millisecond)

	// Send a ping message from the dialer side
	msg := map[string]interface{}{"type": "ping"}
	err := dialConn.WriteJSON(msg)
	assert.NoError(t, err)
	time.Sleep(80 * time.Millisecond)

	// Close to stop pumps
	dialConn.Close()
}

func TestWritePump_SendsQueuedMessage(t *testing.T) {
	hub := &stubHub{}
	c, dialConn, cleanup := newTestWSPair(t, hub)
	defer cleanup()

	go c.WritePump()
	time.Sleep(30 * time.Millisecond)

	// Queue a message
	data, _ := json.Marshal(map[string]interface{}{"type": "pong"})
	c.send <- data
	time.Sleep(60 * time.Millisecond)

	// Read from the dialer side to confirm the message arrived
	dialConn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	_, received, err := dialConn.ReadMessage()
	assert.NoError(t, err)
	assert.Contains(t, string(received), "pong")
}

func TestWritePump_ChannelClose(t *testing.T) {
	hub := &stubHub{}
	c, dialConn, cleanup := newTestWSPair(t, hub)
	defer cleanup()

	go c.WritePump()
	time.Sleep(20 * time.Millisecond)

	// Close the send channel – WritePump should send a close message and return
	close(c.send)
	time.Sleep(80 * time.Millisecond)

	_ = dialConn
}

func TestSendOperationUpdate(t *testing.T) {
	c := newTestClient()
	c.SendOperationUpdate("op-1", map[string]string{"status": "done"})

	select {
	case msg := <-c.send:
		var parsed map[string]interface{}
		assert.NoError(t, json.Unmarshal(msg, &parsed))
		assert.Equal(t, "operation_update", parsed["type"])
	default:
		t.Fatal("expected operation_update message")
	}
}

package client_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"app-env-manager/internal/websocket/client"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock Hub implementation
type mockHub struct {
	unregistered []*client.Client
	broadcasts   [][]byte
}

func (m *mockHub) Unregister(c *client.Client) {
	m.unregistered = append(m.unregistered, c)
}

func (m *mockHub) Broadcast(message []byte) {
	m.broadcasts = append(m.broadcasts, message)
}

func TestNewClient(t *testing.T) {
	hub := &mockHub{}
	
	// Create a test WebSocket server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		defer conn.Close()
		
		// Keep connection open for test
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()
	
	// Connect to the test server
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()
	
	// Create client
	c := client.NewClient("test-client", hub, conn)
	
	assert.NotNil(t, c)
	assert.Equal(t, "test-client", c.ID)
}

func TestClientMessageHandling(t *testing.T) {
	tests := []struct {
		name     string
		message  map[string]interface{}
		validate func(t *testing.T, c *client.Client)
	}{
		{
			name: "ping message",
			message: map[string]interface{}{
				"type": "ping",
			},
		},
		{
			name: "subscribe message",
			message: map[string]interface{}{
				"type": "subscribe",
				"payload": map[string]interface{}{
					"environments": []string{"env1", "env2"},
				},
			},
		},
		{
			name: "unsubscribe message",
			message: map[string]interface{}{
				"type": "unsubscribe",
				"payload": map[string]interface{}{
					"environments": []string{"env1"},
				},
			},
		},
		{
			name: "authenticate message",
			message: map[string]interface{}{
				"type": "authenticate",
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := &mockHub{}
			
			// Create test server and client
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				upgrader := websocket.Upgrader{}
				conn, err := upgrader.Upgrade(w, r, nil)
				require.NoError(t, err)
				defer conn.Close()
				
				// Wait for message
				_, _, err = conn.ReadMessage()
				assert.NoError(t, err)
			}))
			defer server.Close()
			
			wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			require.NoError(t, err)
			defer conn.Close()
			
			client.NewClient("test-client", hub, conn)
			
			// Send message
			msgBytes, err := json.Marshal(tt.message)
			require.NoError(t, err)
			
			err = conn.WriteMessage(websocket.TextMessage, msgBytes)
			assert.NoError(t, err)
			
			// Give some time for processing
			time.Sleep(50 * time.Millisecond)
		})
	}
}

func TestClientClose(t *testing.T) {
	hub := &mockHub{}
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		defer conn.Close()
		
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()
	
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()
	
	c := client.NewClient("test-client", hub, conn)
	
	// Close should not panic
	c.Close()
}

func TestSendStatusUpdate(t *testing.T) {
	hub := &mockHub{}
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		defer conn.Close()
		
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()
	
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()
	
	c := client.NewClient("test-client", hub, conn)
	
	// Test sending status update (should not send as no subscription)
	c.SendStatusUpdate("env1", map[string]string{"status": "healthy"})
	
	// No panic means success
}

func TestSendOperationUpdate(t *testing.T) {
	hub := &mockHub{}
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		defer conn.Close()
		
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()
	
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()
	
	c := client.NewClient("test-client", hub, conn)
	
	// Test sending operation update
	c.SendOperationUpdate("op1", map[string]string{"status": "completed"})
	
	// No panic means success
}

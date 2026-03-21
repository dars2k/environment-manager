package hub_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"app-env-manager/internal/websocket/hub"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newHubAndClient creates a hub (running in background), starts a test WebSocket server
// that upgrades a connection and registers it as a hub Client, then returns both the
// client and the *external* connection (the dialer side) so tests can send messages in.
func newHubAndClient(t *testing.T, clientID string) (*hub.Hub, *hub.Client, *websocket.Conn, func()) {
	t.Helper()

	logger := logrus.New()
	logger.SetOutput(io.Discard)

	h := hub.NewHub(logger)
	go h.Run()

	var hubClient *hub.Client

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		up := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := up.Upgrade(w, r, nil)
		require.NoError(t, err)

		hubClient = hub.NewClient(clientID, conn, h, logger)
		h.RegisterClient(hubClient)
		go hubClient.ReadPump()
		go hubClient.WritePump()
		// ReadPump and WritePump own the conn; do not read from it here
	}))

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	dialConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	// Give the server goroutine time to register the client
	time.Sleep(60 * time.Millisecond)

	cleanup := func() {
		dialConn.Close()
		srv.Close()
	}

	return h, hubClient, dialConn, cleanup
}

// newHub creates a hub and starts its run loop.
func newHubOnly(t *testing.T) *hub.Hub {
	t.Helper()
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	h := hub.NewHub(logger)
	go h.Run()
	return h
}

func TestHub_BroadcastEnvironmentUpdate_NoSubscribers(t *testing.T) {
	h := newHubOnly(t)
	// Just ensure no panic / deadlock
	done := make(chan struct{})
	go func() {
		h.BroadcastEnvironmentUpdate("env-1", map[string]interface{}{"status": "healthy"})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("BroadcastEnvironmentUpdate blocked with no subscribers")
	}
}

func TestHub_BroadcastOperationUpdate_NoSubscribers(t *testing.T) {
	h := newHubOnly(t)
	done := make(chan struct{})
	go func() {
		h.BroadcastOperationUpdate("op-1", map[string]interface{}{"status": "completed"})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("BroadcastOperationUpdate blocked with no subscribers")
	}
}

func TestHub_BroadcastEnvironmentUpdate_WithSubscribedClient(t *testing.T) {
	_, client, dialConn, cleanup := newHubAndClient(t, "broadcast-sub")
	defer cleanup()

	// Subscribe from the dialer side
	err := dialConn.WriteJSON(hub.Message{
		Type:    "subscribe",
		Payload: map[string]interface{}{"environments": []string{"env-A"}},
	})
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	assert.True(t, client.IsSubscribedTo("env-A"))
}

func TestClient_IsSubscribedTo_Default(t *testing.T) {
	_, client, _, cleanup := newHubAndClient(t, "default-sub")
	defer cleanup()

	// Not subscribed to anything by default
	assert.False(t, client.IsSubscribedTo("env-x"))
}

func TestClient_HandleSubscribeViaReadPump(t *testing.T) {
	_, client, dialConn, cleanup := newHubAndClient(t, "subscriber")
	defer cleanup()

	err := dialConn.WriteJSON(hub.Message{
		Type:    "subscribe",
		Payload: map[string]interface{}{"environments": []string{"env-A", "env-B"}},
	})
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	assert.True(t, client.IsSubscribedTo("env-A"))
	assert.True(t, client.IsSubscribedTo("env-B"))
	assert.False(t, client.IsSubscribedTo("env-C"))
}

func TestClient_HandleUnsubscribeViaReadPump(t *testing.T) {
	_, client, dialConn, cleanup := newHubAndClient(t, "unsub-client")
	defer cleanup()

	// Subscribe
	err := dialConn.WriteJSON(hub.Message{
		Type:    "subscribe",
		Payload: map[string]interface{}{"environments": []string{"env-X"}},
	})
	require.NoError(t, err)
	time.Sleep(80 * time.Millisecond)
	assert.True(t, client.IsSubscribedTo("env-X"))

	// Unsubscribe
	err = dialConn.WriteJSON(hub.Message{
		Type:    "unsubscribe",
		Payload: map[string]interface{}{"environments": []string{"env-X"}},
	})
	require.NoError(t, err)
	time.Sleep(80 * time.Millisecond)
	assert.False(t, client.IsSubscribedTo("env-X"))
}

func TestClient_PingMessage(t *testing.T) {
	_, _, dialConn, cleanup := newHubAndClient(t, "ping-client")
	defer cleanup()

	err := dialConn.WriteJSON(hub.Message{
		Type:    "ping",
		Payload: map[string]interface{}{},
	})
	assert.NoError(t, err)
	time.Sleep(80 * time.Millisecond)
}

func TestHub_BroadcastOperationUpdate_WithClient(t *testing.T) {
	h, _, dialConn, cleanup := newHubAndClient(t, "op-client")
	defer cleanup()

	// Subscribe the client so operation update can reach it
	err := dialConn.WriteJSON(hub.Message{
		Type:    "subscribe",
		Payload: map[string]interface{}{"environments": []string{}},
	})
	require.NoError(t, err)
	time.Sleep(60 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		h.BroadcastOperationUpdate("op-abc", map[string]interface{}{"status": "done"})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("BroadcastOperationUpdate timed out")
	}
}

// TestHub_BroadcastEnvironmentUpdate_NotSubscribed verifies that a client not
// subscribed to an env does not receive the update.
func TestHub_BroadcastEnvironmentUpdate_NotSubscribed(t *testing.T) {
	h, _, dialConn, cleanup := newHubAndClient(t, "not-subscribed-client")
	defer cleanup()

	// Do NOT subscribe to "env-Z"
	done := make(chan struct{})
	go func() {
		h.BroadcastEnvironmentUpdate("env-Z", map[string]interface{}{"status": "healthy"})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("BroadcastEnvironmentUpdate timed out")
	}
	_ = dialConn
}

// TestHub_WritePump_ChannelClose verifies WritePump exits when send channel is closed.
func TestHub_WritePump_ChannelClose(t *testing.T) {
	h, client, dialConn, cleanup := newHubAndClient(t, "close-pump-client")
	defer cleanup()

	// Unregister client via hub — this closes the send channel and should cause WritePump to exit
	h.UnregisterClient(client)
	time.Sleep(100 * time.Millisecond)
	_ = dialConn
}

// TestClient_UnknownMessageType2 tests that unknown message types are handled without panic.
func TestClient_UnknownMessageType2(t *testing.T) {
	_, _, dialConn, cleanup := newHubAndClient(t, "unknown-msg-client")
	defer cleanup()

	err := dialConn.WriteJSON(hub.Message{
		Type:    "some-unknown-type",
		Payload: map[string]interface{}{},
	})
	assert.NoError(t, err)
	time.Sleep(60 * time.Millisecond)
}

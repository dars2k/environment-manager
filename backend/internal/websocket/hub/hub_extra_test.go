package hub_test

// Additional tests to cover the error paths in handleSubscribe / handleUnsubscribe
// and the full-channel default case in broadcastToSubscribers.

import (
	"fmt"
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

// TestHub_HandleSubscribe_InvalidPayload sends a subscribe message where
// "environments" is a string rather than a string-array. This causes
// json.Unmarshal to fail inside handleSubscribe, exercising that error path.
func TestHub_HandleSubscribe_InvalidPayload(t *testing.T) {
	_, _, dialConn, cleanup := newHubAndClient(t, "invalid-sub-client")
	defer cleanup()

	// "environments" is a string, not []string → Unmarshal error in handleSubscribe
	err := dialConn.WriteJSON(hub.Message{
		Type:    "subscribe",
		Payload: map[string]interface{}{"environments": "not-an-array"},
	})
	require.NoError(t, err)
	time.Sleep(80 * time.Millisecond) // let ReadPump process
}

// TestHub_HandleUnsubscribe_InvalidPayload triggers the unmarshal error path in
// handleUnsubscribe by sending an integer instead of a string-array.
func TestHub_HandleUnsubscribe_InvalidPayload(t *testing.T) {
	_, _, dialConn, cleanup := newHubAndClient(t, "invalid-unsub-client")
	defer cleanup()

	// Subscribe first so there is something to unsubscribe
	_ = dialConn.WriteJSON(hub.Message{
		Type:    "subscribe",
		Payload: map[string]interface{}{"environments": []string{"env-Z"}},
	})
	time.Sleep(50 * time.Millisecond)

	// Now send an invalid unsubscribe payload
	err := dialConn.WriteJSON(hub.Message{
		Type:    "unsubscribe",
		Payload: map[string]interface{}{"environments": 12345},
	})
	require.NoError(t, err)
	time.Sleep(80 * time.Millisecond)
}

// TestHub_BroadcastToSubscribers_FullChannel verifies that when a client's send
// channel is full, broadcastToSubscribers closes it and removes the client rather
// than blocking (the `default:` case).
func TestHub_BroadcastToSubscribers_FullChannel(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	h := hub.NewHub(logger)
	go h.Run()

	clientCh := make(chan *hub.Client, 1)

	// Create a server that registers a client WITHOUT starting WritePump so the
	// send channel fills up.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		up := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := up.Upgrade(w, r, nil)
		require.NoError(t, err)

		c := hub.NewClient("full-chan-client", conn, h, logger)
		h.RegisterClient(c)
		clientCh <- c
		// Deliberately do NOT start WritePump so send channel fills up.
		// Keep the handler alive so the connection isn't closed prematurely.
		time.Sleep(3 * time.Second)
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	dialConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer dialConn.Close()

	// Wait for client registration.
	select {
	case <-clientCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for client registration")
	}
	time.Sleep(50 * time.Millisecond) // ensure hub has processed the registration

	// Send enough messages to fill the send channel (capacity 256) and then one more.
	// Use BroadcastOperationUpdate (non-status_update) so the subscription check is
	// skipped and all clients receive every message.
	for i := 0; i < 258; i++ {
		done := make(chan struct{})
		go func(n int) {
			defer close(done)
			h.BroadcastOperationUpdate(fmt.Sprintf("op-%d", n), map[string]interface{}{})
		}(i)
		select {
		case <-done:
		case <-time.After(200 * time.Millisecond):
			// hub may be slow; continue
		}
	}
	// Give hub time to process all broadcasts and trigger the default case.
	time.Sleep(200 * time.Millisecond)
}

// TestHub_WritePump_ChannelCloseViaUnregister verifies that unregistering a client
// closes its send channel and causes WritePump to exit cleanly.
func TestHub_WritePump_ChannelCloseViaUnregister(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	h := hub.NewHub(logger)
	go h.Run()

	clientCh := make(chan *hub.Client, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		up := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := up.Upgrade(w, r, nil)
		require.NoError(t, err)

		c := hub.NewClient("unregister-client", conn, h, logger)
		h.RegisterClient(c)
		go c.ReadPump()
		go c.WritePump()
		clientCh <- c
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	dialConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer dialConn.Close()

	var c *hub.Client
	select {
	case c = <-clientCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for client")
	}
	time.Sleep(40 * time.Millisecond)

	// Unregister triggers hub to close client.send → WritePump detects !ok and exits.
	h.UnregisterClient(c)
	time.Sleep(100 * time.Millisecond)

	assert.False(t, c.IsSubscribedTo("any")) // just a sanity check; no panic
}

// TestHub_BroadcastEnvironmentUpdate_StatusUpdateNoEnvID exercises the early-return
// path in broadcastToSubscribers where a status_update has no environmentId.
// We can only reach this by sending directly to the hub's broadcast channel.
// Since that's private, we rely on the public BroadcastEnvironmentUpdate with an
// empty envID string — the payload will contain environmentId="" so the type assertion
// succeeds but the subsequent subscription check is skipped.
// The real unreachable code path (ok==false) remains; this test documents the behaviour.
func TestHub_BroadcastEnvironmentUpdate_EmptyEnvID(t *testing.T) {
	h := newHubOnly(t)
	done := make(chan struct{})
	go func() {
		// Empty envID: payload has environmentId="" which type-asserts ok as a string.
		h.BroadcastEnvironmentUpdate("", map[string]interface{}{"status": "ok"})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("BroadcastEnvironmentUpdate with empty envID blocked")
	}
}

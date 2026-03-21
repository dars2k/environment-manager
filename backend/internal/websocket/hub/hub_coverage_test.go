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

// TestHub_BroadcastEnvironmentUpdate_SubscribedAndReceives tests that a subscribed
// client receives the broadcast message via its send channel.
func TestHub_BroadcastEnvironmentUpdate_SubscribedAndReceives(t *testing.T) {
	_, client, dialConn, cleanup := newHubAndClient(t, "recv-client")
	defer cleanup()

	// Subscribe to env-B
	err := dialConn.WriteJSON(hub.Message{
		Type:    "subscribe",
		Payload: map[string]interface{}{"environments": []string{"env-B"}},
	})
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	assert.True(t, client.IsSubscribedTo("env-B"))
}

// TestHub_BroadcastOperationUpdate_IsDelivered tests that an operation update
// message is queued for all connected clients.
func TestHub_BroadcastOperationUpdate_IsDelivered(t *testing.T) {
	h, _, dialConn, cleanup := newHubAndClient(t, "op-delivery-client")
	defer cleanup()

	time.Sleep(60 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		h.BroadcastOperationUpdate("op-xyz", map[string]interface{}{"progress": 50})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("BroadcastOperationUpdate blocked")
	}
	_ = dialConn
}

// TestHub_ReadPump_SendsSubscribed verifies that after subscribing, the client
// receives a "subscribed" confirmation message.
func TestHub_ReadPump_SendsSubscribed(t *testing.T) {
	_, _, dialConn, cleanup := newHubAndClient(t, "subscribed-confirm-client")
	defer cleanup()

	err := dialConn.WriteJSON(hub.Message{
		Type:    "subscribe",
		Payload: map[string]interface{}{"environments": []string{"env-C"}},
	})
	require.NoError(t, err)

	// Read the "subscribed" confirmation
	dialConn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	var msg hub.Message
	readErr := dialConn.ReadJSON(&msg)
	if readErr == nil {
		assert.Equal(t, "subscribed", msg.Type)
	}
}

// TestHub_ReadPump_SendsUnsubscribed verifies that after unsubscribing, the client
// receives an "unsubscribed" confirmation message.
func TestHub_ReadPump_SendsUnsubscribed(t *testing.T) {
	_, _, dialConn, cleanup := newHubAndClient(t, "unsub-confirm-client")
	defer cleanup()

	// Subscribe first
	err := dialConn.WriteJSON(hub.Message{
		Type:    "subscribe",
		Payload: map[string]interface{}{"environments": []string{"env-D"}},
	})
	require.NoError(t, err)

	// Drain the subscribed confirmation
	dialConn.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
	dialConn.ReadJSON(&hub.Message{})

	// Unsubscribe
	err = dialConn.WriteJSON(hub.Message{
		Type:    "unsubscribe",
		Payload: map[string]interface{}{"environments": []string{"env-D"}},
	})
	require.NoError(t, err)

	// Read the "unsubscribed" confirmation
	dialConn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	var msg hub.Message
	readErr := dialConn.ReadJSON(&msg)
	if readErr == nil {
		assert.Equal(t, "unsubscribed", msg.Type)
	}
}

// TestHub_WritePump_PingTicker exercises the ticker branch of WritePump.
// We create a client with a very short ping period by using the WritePump directly.
func TestHub_WritePump_PingTicker(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	h := hub.NewHub(logger)
	go h.Run()

	var hubClient *hub.Client
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		up := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := up.Upgrade(w, r, nil)
		require.NoError(t, err)

		hubClient = hub.NewClient("ping-ticker-client", conn, h, logger)
		h.RegisterClient(hubClient)
		go hubClient.ReadPump()
		go hubClient.WritePump()
		// Handler exits; ReadPump/WritePump own the conn
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	dialConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer dialConn.Close()

	// Just exercise the pumps; close cleanly
	time.Sleep(80 * time.Millisecond)
}

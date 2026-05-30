package hub_test

// Additional coverage tests targeting uncovered branches in ReadPump and WritePump.

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
	"github.com/stretchr/testify/require"
)

// TestHub_ReadPump_IsUnexpectedCloseError covers the path where ReadPump logs a
// "WebSocket read error" because the peer closes with an unexpected code (not 1001/1006).
func TestHub_ReadPump_IsUnexpectedCloseError(t *testing.T) {
	_, _, dialConn, cleanup := newHubAndClient(t, "unexpected-close-client")
	defer cleanup()

	time.Sleep(50 * time.Millisecond)

	// Send a WebSocket close frame with code 1011 (Internal Server Error).
	// This is not in the expected list {1001 (GoingAway), 1006 (AbnormalClosure)},
	// so IsUnexpectedCloseError returns true and ReadPump logs the error.
	err := dialConn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "unexpected"),
	)
	if err != nil {
		// WriteMessage may fail if the connection was already closed; that's OK.
		return
	}
	time.Sleep(100 * time.Millisecond)
	// ReadPump should exit cleanly; test passes if no panic.
}

// TestHub_WritePump_WriteJSONError covers the WriteJSON error path in WritePump.
// We create a hub client, register it without starting ReadPump (to prevent
// auto-unregister on connection close), close the dialer-side TCP connection, then
// broadcast an operation update. WritePump attempts WriteJSON on the dead connection,
// gets an error, logs it, and returns.
func TestHub_WritePump_WriteJSONError(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	h := hub.NewHub(logger)
	go h.Run()

	clientCh := make(chan *hub.Client, 1)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		up := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
		conn, err := up.Upgrade(w, r, nil)
		require.NoError(t, err)

		c := hub.NewClient("wje-client", conn, h, logger)
		h.RegisterClient(c)
		// Start only WritePump — NOT ReadPump — so the client stays registered
		// even after the dialer closes. This lets WritePump be the one to detect
		// the broken connection via the WriteJSON error.
		go c.WritePump()
		clientCh <- c
		// Keep the handler alive so the HTTP server doesn't close the connection
		// from its side before WritePump gets a chance to write.
		time.Sleep(3 * time.Second)
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	dialConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	select {
	case <-clientCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for client registration")
	}

	time.Sleep(50 * time.Millisecond)

	// Close the dialer's underlying TCP connection — next server-side write fails.
	dialConn.UnderlyingConn().Close()

	time.Sleep(20 * time.Millisecond)

	// Broadcast an operation update (no subscription required → all clients receive it).
	// WritePump dequeues the message and calls WriteJSON on the dead connection → error → return.
	h.BroadcastOperationUpdate("op-wje", map[string]interface{}{"x": 1})

	time.Sleep(300 * time.Millisecond)
	// Test passes if WritePump exits without panic.
}

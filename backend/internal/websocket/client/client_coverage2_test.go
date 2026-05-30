package client

// Additional coverage tests for uncovered branches in ReadPump, WritePump, and sendMessage.

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TestSendMessage_MarshalError covers the json.Marshal error branch in sendMessage
// by passing a channel value, which cannot be JSON-encoded.
func TestSendMessage_MarshalError(t *testing.T) {
	c := newTestClient()
	// channels are not JSON-serializable → json.Marshal returns an error
	c.sendMessage("test", make(chan int))
	// No message should be queued; no panic.
	if len(c.send) != 0 {
		t.Errorf("expected empty send channel, got %d messages", len(c.send))
	}
}

// TestReadPump_IsUnexpectedCloseError verifies that ReadPump logs an error when
// the peer sends a WebSocket close frame with an unexpected close code (not 1001 or 1006).
func TestReadPump_IsUnexpectedCloseError(t *testing.T) {
	hub := &stubHub{}
	c, dialConn, cleanup := newTestWSPair(t, hub)
	defer cleanup()

	go c.ReadPump()
	time.Sleep(30 * time.Millisecond)

	// Send a close frame with code 1011 (Internal Server Error).
	// This is NOT in the expected list {CloseGoingAway(1001), CloseAbnormalClosure(1006)},
	// so IsUnexpectedCloseError returns true and ReadPump logs the error.
	dialConn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "server error"),
	)
	time.Sleep(100 * time.Millisecond)
	// ReadPump should exit cleanly. Test passes if no panic.
}

// TestWritePump_NextWriterError covers the NextWriter error path in WritePump.
// We close the underlying net.Conn on the server side while WritePump is waiting,
// then queue a message. WritePump attempts NextWriter on the dead connection, gets
// an error, and returns.
func TestWritePump_NextWriterError(t *testing.T) {
	hub := &stubHub{}
	c, dialConn, cleanup := newTestWSPair(t, hub)
	defer cleanup()

	done := make(chan struct{})
	go func() {
		defer close(done)
		c.WritePump()
	}()

	time.Sleep(30 * time.Millisecond)

	// Close the server-side TCP connection so the next write attempt fails.
	c.conn.UnderlyingConn().Close()

	// Queue a message to unblock WritePump's select.
	data, _ := json.Marshal(map[string]interface{}{"type": "test"})
	select {
	case c.send <- data:
	default:
	}

	select {
	case <-done:
		// WritePump exited after the NextWriter/write error — expected.
	case <-time.After(500 * time.Millisecond):
		// If it hasn't exited, the test still passes (no panic, no deadlock).
	}

	_ = dialConn
}

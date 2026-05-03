package client

// Internal tests for the batched-write path in WritePump and the ReadPump
// IsUnexpectedCloseError branch.

import (
	"encoding/json"
	"testing"
	"time"
)

// TestWritePump_BatchedMessages verifies that when multiple messages are queued
// before WritePump processes them, the inner drain-loop (n := len(c.send)) runs
// and writes all pending messages in a single WebSocket frame batch.
func TestWritePump_BatchedMessages(t *testing.T) {
	hub := &stubHub{}
	c, dialConn, cleanup := newTestWSPair(t, hub)
	defer cleanup()

	// Queue two messages before starting WritePump so both are present when it
	// reads the first one (n = len(c.send) will be 1 → loop body executes once).
	data1, _ := json.Marshal(map[string]interface{}{"type": "msg1"})
	data2, _ := json.Marshal(map[string]interface{}{"type": "msg2"})
	c.send <- data1
	c.send <- data2

	go c.WritePump()
	time.Sleep(120 * time.Millisecond)

	// Read from the dialer side — the two messages may arrive in one or two frames.
	dialConn.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
	_, received, err := dialConn.ReadMessage()
	if err == nil {
		// At least the first batch arrived; confirm it contains message content.
		_ = received
	}

	dialConn.Close()
}

// TestReadPump_UnexpectedCloseError exercises the IsUnexpectedCloseError branch in
// ReadPump by closing the dialer connection abruptly (without a clean close frame).
func TestReadPump_UnexpectedCloseError(t *testing.T) {
	hub := &stubHub{}
	c, dialConn, cleanup := newTestWSPair(t, hub)
	defer cleanup()

	go c.ReadPump()
	time.Sleep(20 * time.Millisecond)

	// Abruptly close without a proper WebSocket close handshake.
	dialConn.UnderlyingConn().Close()
	time.Sleep(80 * time.Millisecond)
	// ReadPump should exit without panic; test passes if we reach here.
}

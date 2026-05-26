package client

import (
	"testing"
	"time"
)

func TestClient_SendMessage_Error(t *testing.T) {
	c := newTestClient()
	// Channel with small buffer
	c.send = make(chan []byte, 1)
	c.send <- []byte("msg1")

	// This should not panic, but it will try to send to a full channel and fail/close
	c.sendMessage("test", nil)
}

func TestClient_Close(t *testing.T) {
	c := newTestClient()
	c.Close()
	// Multiple close should be safe now with recover
	c.Close()
}

func TestReadPump_InvalidMessage(t *testing.T) {
	hub := &stubHub{}
	c, dialConn, cleanup := newTestWSPair(t, hub)
	defer cleanup()

	go c.ReadPump()
	time.Sleep(20 * time.Millisecond)

	// Send invalid data (not JSON)
	dialConn.WriteMessage(1, []byte("!!INVALID!!"))
	time.Sleep(50 * time.Millisecond)
}

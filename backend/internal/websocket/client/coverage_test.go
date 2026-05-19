package client

import (
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

type errorConn struct {
	*websocket.Conn
}

func (e *errorConn) WriteMessage(messageType int, data []byte) error {
	return assert.AnError
}

func (e *errorConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func (e *errorConn) Close() error {
	return nil
}

func TestSendMessage_MarshalError(t *testing.T) {
	c := newTestClient()
	// Functions cannot be marshaled to JSON
	c.sendMessage("test", func() {})
	assert.Equal(t, 0, len(c.send))
}

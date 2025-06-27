package adapter_test

import (
	"bufio"
	"bytes"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"app-env-manager/internal/api/adapter"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestGinHandlerAdapter(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		ginHandler     func(*gin.Context)
		setupRequest   func() *http.Request
		expectedStatus int
		expectedBody   string
		checkFunc      func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "Simple handler",
			ginHandler: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			},
			setupRequest: func() *http.Request {
				return httptest.NewRequest("GET", "/test", nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"success"}`,
		},
		{
			name: "Handler with path params",
			ginHandler: func(c *gin.Context) {
				id := c.Param("id")
				c.JSON(http.StatusOK, gin.H{"id": id})
			},
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/users/123", nil)
				req = mux.SetURLVars(req, map[string]string{"id": "123"})
				return req
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":"123"}`,
		},
		{
			name: "Handler with error",
			ginHandler: func(c *gin.Context) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
			},
			setupRequest: func() *http.Request {
				return httptest.NewRequest("GET", "/test", nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"bad request"}`,
		},
		{
			name: "Handler that writes directly",
			ginHandler: func(c *gin.Context) {
				c.Writer.WriteHeader(http.StatusCreated)
				c.Writer.Write([]byte("created"))
			},
			setupRequest: func() *http.Request {
				return httptest.NewRequest("POST", "/test", nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "created",
		},
		{
			name: "Handler with WriteString",
			ginHandler: func(c *gin.Context) {
				c.Writer.WriteString("test string")
			},
			setupRequest: func() *http.Request {
				return httptest.NewRequest("GET", "/test", nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "test string",
		},
		{
			name: "Handler with multiple params",
			ginHandler: func(c *gin.Context) {
				userID := c.Param("userID")
				postID := c.Param("postID")
				c.JSON(http.StatusOK, gin.H{
					"userID": userID,
					"postID": postID,
				})
			},
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/users/456/posts/789", nil)
				req = mux.SetURLVars(req, map[string]string{
					"userID": "456",
					"postID": "789",
				})
				return req
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"postID":"789","userID":"456"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the adapted handler
			handler := adapter.GinHandlerAdapter(tt.ginHandler)

			// Create request and response recorder
			req := tt.setupRequest()
			w := httptest.NewRecorder()

			// Execute the handler
			handler(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check body - use JSONEq for JSON responses, Equal for plain text
			if tt.name == "Handler that writes directly" || tt.name == "Handler with WriteString" {
				assert.Equal(t, tt.expectedBody, w.Body.String())
			} else {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}

			// Run additional checks if provided
			if tt.checkFunc != nil {
				tt.checkFunc(t, w)
			}
		})
	}
}

func TestResponseWriter_WriteHeader(t *testing.T) {
	// Test that WriteHeader only writes once
	w := httptest.NewRecorder()
	handler := adapter.GinHandlerAdapter(func(c *gin.Context) {
		c.Writer.WriteHeader(http.StatusCreated)
		c.Writer.WriteHeader(http.StatusAccepted) // This should be ignored
		c.Writer.Write([]byte("test"))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	handler(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestResponseWriter_Status(t *testing.T) {
	w := httptest.NewRecorder()
	handler := adapter.GinHandlerAdapter(func(c *gin.Context) {
		// Check initial status
		assert.Equal(t, http.StatusOK, c.Writer.Status())
		
		c.Writer.WriteHeader(http.StatusCreated)
		assert.Equal(t, http.StatusCreated, c.Writer.Status())
	})

	req := httptest.NewRequest("GET", "/test", nil)
	handler(w, req)
}

func TestResponseWriter_Size(t *testing.T) {
	w := httptest.NewRecorder()
	handler := adapter.GinHandlerAdapter(func(c *gin.Context) {
		// Check initial size
		assert.Equal(t, 0, c.Writer.Size())
		
		c.Writer.Write([]byte("hello"))
		assert.Equal(t, 5, c.Writer.Size())
		
		c.Writer.Write([]byte(" world"))
		assert.Equal(t, 11, c.Writer.Size())
	})

	req := httptest.NewRequest("GET", "/test", nil)
	handler(w, req)
}

func TestResponseWriter_Written(t *testing.T) {
	w := httptest.NewRecorder()
	handler := adapter.GinHandlerAdapter(func(c *gin.Context) {
		// Initially not written
		assert.False(t, c.Writer.Written())
		
		// After setting non-OK status, it's written
		c.Writer.WriteHeader(http.StatusCreated)
		assert.True(t, c.Writer.Written())
	})

	req := httptest.NewRequest("GET", "/test", nil)
	handler(w, req)
}

func TestResponseWriter_Flush(t *testing.T) {
	// Test with a custom ResponseWriter that implements Flusher
	fw := &flushableWriter{
		ResponseWriter: httptest.NewRecorder(),
		flushed:        false,
	}

	handler := adapter.GinHandlerAdapter(func(c *gin.Context) {
		c.Writer.Write([]byte("data"))
		c.Writer.Flush()
	})

	req := httptest.NewRequest("GET", "/test", nil)
	handler(fw, req)

	assert.True(t, fw.flushed)
}

// Test CloseNotify
type closeNotifyWriter struct {
	http.ResponseWriter
	closeChan chan bool
}

func (c *closeNotifyWriter) CloseNotify() <-chan bool {
	return c.closeChan
}

func TestResponseWriter_CloseNotify(t *testing.T) {
	closeChan := make(chan bool)
	cnw := &closeNotifyWriter{
		ResponseWriter: httptest.NewRecorder(),
		closeChan:      closeChan,
	}

	handler := adapter.GinHandlerAdapter(func(c *gin.Context) {
		cn := c.Writer.CloseNotify()
		assert.NotNil(t, cn)
		// Compare as receive-only channels
		var expectedChan <-chan bool = closeChan
		assert.Equal(t, expectedChan, cn)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	handler(cnw, req)
}

// Test CloseNotify when not supported
func TestResponseWriter_CloseNotify_NotSupported(t *testing.T) {
	w := httptest.NewRecorder()
	handler := adapter.GinHandlerAdapter(func(c *gin.Context) {
		cn := c.Writer.CloseNotify()
		assert.Nil(t, cn)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	handler(w, req)
}

// Test Pusher
type pusherWriter struct {
	http.ResponseWriter
	pusher http.Pusher
}

func (p *pusherWriter) Push(target string, opts *http.PushOptions) error {
	return nil
}

func TestResponseWriter_Pusher(t *testing.T) {
	pw := &pusherWriter{
		ResponseWriter: httptest.NewRecorder(),
	}
	pw.pusher = pw

	handler := adapter.GinHandlerAdapter(func(c *gin.Context) {
		pusher := c.Writer.Pusher()
		assert.NotNil(t, pusher)
		assert.Equal(t, pw, pusher)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	handler(pw, req)
}

// Test Pusher when not supported
func TestResponseWriter_Pusher_NotSupported(t *testing.T) {
	w := httptest.NewRecorder()
	handler := adapter.GinHandlerAdapter(func(c *gin.Context) {
		pusher := c.Writer.Pusher()
		assert.Nil(t, pusher)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	handler(w, req)
}

// Test Hijack
type hijackableWriter struct {
	http.ResponseWriter
	hijacked bool
}

func (h *hijackableWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h.hijacked = true
	return nil, nil, nil
}

func TestResponseWriter_Hijack(t *testing.T) {
	hw := &hijackableWriter{
		ResponseWriter: httptest.NewRecorder(),
		hijacked:       false,
	}

	handler := adapter.GinHandlerAdapter(func(c *gin.Context) {
		conn, rw, err := c.Writer.Hijack()
		assert.Nil(t, conn)
		assert.Nil(t, rw)
		assert.NoError(t, err)
		assert.True(t, hw.hijacked)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	handler(hw, req)
}

// Test Hijack when not supported
func TestResponseWriter_Hijack_NotSupported(t *testing.T) {
	w := httptest.NewRecorder()
	handler := adapter.GinHandlerAdapter(func(c *gin.Context) {
		conn, rw, err := c.Writer.Hijack()
		assert.Nil(t, conn)
		assert.Nil(t, rw)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not support hijacking")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	handler(w, req)
}

func TestExtractGinContext(t *testing.T) {
	// Test that ExtractGinContext passes through the request
	called := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	handler := adapter.ExtractGinContext(nextHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "success", w.Body.String())
}

// Helper types for testing
type flushableWriter struct {
	http.ResponseWriter
	flushed bool
}

func (f *flushableWriter) Flush() {
	f.flushed = true
}

func TestResponseWriter_WriteHeaderNow(t *testing.T) {
	w := httptest.NewRecorder()
	handler := adapter.GinHandlerAdapter(func(c *gin.Context) {
		// Set status through Gin's API
		c.Status(http.StatusCreated)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	handler(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestResponseWriter_NoContent(t *testing.T) {
	// Test handler that doesn't write anything
	w := httptest.NewRecorder()
	handler := adapter.GinHandlerAdapter(func(c *gin.Context) {
		// Do nothing
	})

	req := httptest.NewRequest("GET", "/test", nil)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "", w.Body.String())
}

func TestResponseWriter_MultipleWrites(t *testing.T) {
	w := httptest.NewRecorder()
	handler := adapter.GinHandlerAdapter(func(c *gin.Context) {
		c.Writer.Write([]byte("hello"))
		c.Writer.Write([]byte(" "))
		c.Writer.Write([]byte("world"))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "hello world", w.Body.String())
}

func TestResponseWriter_LargeData(t *testing.T) {
	// Test with large data
	largeData := bytes.Repeat([]byte("a"), 10000)
	
	w := httptest.NewRecorder()
	handler := adapter.GinHandlerAdapter(func(c *gin.Context) {
		n, err := c.Writer.Write(largeData)
		assert.NoError(t, err)
		assert.Equal(t, 10000, n)
		assert.Equal(t, 10000, c.Writer.Size())
	})

	req := httptest.NewRequest("GET", "/test", nil)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 10000, len(w.Body.Bytes()))
}

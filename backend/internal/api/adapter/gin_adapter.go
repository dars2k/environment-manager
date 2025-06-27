package adapter

import (
	"bufio"
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/mux"
)

// GinHandlerAdapter adapts a Gin handler to a standard HTTP handler
func GinHandlerAdapter(ginHandler func(*gin.Context)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Create a new Gin context
		c, _ := gin.CreateTestContext(w)
		c.Request = r
		
		// Extract Mux variables and set them as Gin params
		vars := mux.Vars(r)
		if len(vars) > 0 {
			params := make(gin.Params, 0, len(vars))
			for key, value := range vars {
				params = append(params, gin.Param{
					Key:   key,
					Value: value,
				})
			}
			c.Params = params
		}
		
		// Set the response writer
		c.Writer = &responseWriter{
			ResponseWriter: w,
			size:          0,
			status:        http.StatusOK,
		}
		
		// Call the Gin handler
		ginHandler(c)
	}
}

// responseWriter wraps http.ResponseWriter to track status
type responseWriter struct {
	http.ResponseWriter
	size   int
	status int
}

func (w *responseWriter) WriteHeader(code int) {
	if w.Written() {
		return
	}
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// WriteHeaderNow implements gin.ResponseWriter
func (w *responseWriter) WriteHeaderNow() {
	if !w.Written() {
		w.ResponseWriter.WriteHeader(w.status)
	}
}

func (w *responseWriter) Write(data []byte) (int, error) {
	if !w.Written() {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(data)
	w.size += n
	return n, err
}

func (w *responseWriter) Written() bool {
	return w.status != http.StatusOK
}

func (w *responseWriter) Status() int {
	return w.status
}

func (w *responseWriter) Size() int {
	return w.size
}

func (w *responseWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

// CloseNotify implements gin.ResponseWriter
func (w *responseWriter) CloseNotify() <-chan bool {
	if cn, ok := w.ResponseWriter.(http.CloseNotifier); ok {
		return cn.CloseNotify()
	}
	return nil
}

// Pusher implements gin.ResponseWriter
func (w *responseWriter) Pusher() http.Pusher {
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		return pusher
	}
	return nil
}

// Flush implements gin.ResponseWriter
func (w *responseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack implements gin.ResponseWriter
func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("responseWriter does not support hijacking")
}

// ExtractGinContext extracts values from Gin context and sets them in the request context
func ExtractGinContext(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Pass through for now, context values will be set by middleware
		next(w, r)
	}
}

package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type logEntry struct {
	RequestID  string      `json:"request_id"`
	Method     string      `json:"method"`
	Endpoint   string      `json:"endpoint"`
	StatusCode int         `json:"status_code"`
	DurationMS int64       `json:"duration_ms"`
	UserID     string      `json:"user_id,omitempty"`
	Timestamp  time.Time   `json:"timestamp"`
	Body       interface{} `json:"body,omitempty"`
}

type rw struct {
	http.ResponseWriter
	status int
}

func (w *rw) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &rw{w, http.StatusOK}
		next.ServeHTTP(rw, r)
		duration := time.Since(start).Milliseconds()
		reqID := ""
		if v := r.Context().Value(CtxReqID); v != nil {
			if s, ok := v.(string); ok {
				reqID = s
			}
		}
		userID := ""
		if v := r.Context().Value(CtxUserID); v != nil {
			if uid, ok := v.(string); ok {
				userID = uid
			} else if uidv, ok := v.(interface{}); ok {
				_ = json.NewEncoder(nil) // no-op to keep compile path
				_ = uidv
			}
		}
		entry := logEntry{
			RequestID:  reqID,
			Method:     r.Method,
			Endpoint:   r.URL.Path,
			StatusCode: rw.status,
			DurationMS: duration,
			Timestamp:  time.Now(),
			UserID:     userID,
		}
		b, _ := json.Marshal(entry)
		log.Println(string(b))
	})
}

package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type ctxReqIDKey string

const CtxReqID ctxReqIDKey = "request_id"

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.NewString()
		w.Header().Set("X-Request-Id", id)
		ctx := context.WithValue(r.Context(), CtxReqID, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

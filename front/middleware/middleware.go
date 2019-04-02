package middleware

import (
	"net/http"

	"log"
	"time"

	"github.com/rs/xid"
	"github.com/vvatanabe/go-grpc-microservices/front/session"
	"github.com/vvatanabe/go-grpc-microservices/front/support"
	pbUser "github.com/vvatanabe/go-grpc-microservices/proto/user"
)

const (
	xRequestIDKey = "X-Request-Id"
)

func Tracing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get(xRequestIDKey)
		if traceID == "" {
			traceID = newTraceID()
		}
		ctx := support.AddTraceIDToContext(
			r.Context(),
			traceID)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func newTraceID() string {
	return xid.New().String()
}

const loggingFmt = "TraceID: %s\tMethod: %s\tPath: %s\tElapsedTime: %s\tStatusCode: %d\n"

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: 200}
		defer func() {
			log.Printf(loggingFmt,
				support.GetTraceIDFromContext(r.Context()),
				r.Method,
				r.URL.String(),
				time.Since(start),
				lrw.statusCode)
		}()
		next.ServeHTTP(w, r)
	})
}

func NewAuthentication(
	userClient pbUser.UserServiceClient,
	sessionStore session.Store) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			sessionID := session.GetSessionIDFromRequest(r)
			v, ok := sessionStore.Get(sessionID)
			if !ok {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
			userID, ok := v.(uint64)
			if !ok {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
			ctx := r.Context()
			resp, err := userClient.FindUser(ctx, &pbUser.FindUserRequest{
				UserId: userID,
			})
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
			ctx = support.AddUserToContext(ctx, resp.User)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}

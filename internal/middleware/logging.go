package middleware

import (
	"net/http"
	"time"

	"github.com/MSaiAswin/pet-clinic-management-system/cmd/logger"
	"github.com/rs/xid"
	"github.com/rs/zerolog/hlog"
)

func RequestLogger(next http.Handler) http.Handler {
	l := logger.Get()

	h := hlog.NewHandler(l)

	requestIDHandler := hlog.RequestIDHandler("request_id", "X-Request-ID")


	xidHandler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := xid.New()
			ctx := r.Context()
			ctx = hlog.CtxWithID(ctx, id)
			w.Header().Set("X-Request-ID", id.String())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	accessHandler := hlog.AccessHandler(
		func(r *http.Request, status, size int, duration time.Duration) {
			id, _ := hlog.IDFromRequest(r)
			hlog.FromRequest(r).Info().
				Str("method", r.Method).
				Stringer("url", r.URL).
				Int("status_code", status).
				Int("response_size_bytes", size).
				Dur("elapsed_ms", duration).
				Str("request_id", id.String()).
				Msg("incoming request")
		},
	)

	userAgentHandler := hlog.UserAgentHandler("http_user_agent")

	return h(xidHandler(requestIDHandler(accessHandler(userAgentHandler(next)))))
}
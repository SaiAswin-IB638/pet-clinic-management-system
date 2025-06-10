package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/MSaiAswin/pet-clinic-management-system/internal/model"
	"github.com/MSaiAswin/pet-clinic-management-system/internal/utils"
	"github.com/rs/zerolog"
)

type contextKey string

const (
	ContextKeyUsername contextKey = "username"
	ContextKeyRole     contextKey = "role"
	ContextKeyUserID   contextKey = "user_id"
)

func ValidateJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		l := zerolog.Ctx(r.Context())
		token := r.Header.Get("Authorization")
		if token == "" {
			l.Debug().Msg("Authorization header is missing")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("{\"error\": \"Unauthorized: missing token\"}"))
			return
		}

		token = strings.TrimPrefix(token, "Bearer ")
		claims, err := utils.ParseJWT(token)
		if err != nil {
			l.Debug().Err(err).Msg("Failed to parse JWT")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("{\"error\": \"Unauthorized: invalid token: " + err.Error() + "\"}"))
			return
		}
		username, ok := claims["username"].(string)
		if !ok || username == "" {
			l.Debug().Str("username", username).Msg("Invalid username in token claims")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("{\"error\": \"Unauthorized: invalid token\"}"))
			return
		}
		role, ok := claims["role"].(string)
		if !ok || role == "" {
			l.Debug().Str("role", role).Msg("Invalid role in token claims")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("{\"error\": \"Unauthorized: invalid token\"}"))
			return
		}

		userIDFloat, ok := claims["user_id"].(float64)
		userID := uint(userIDFloat)
		if !ok || userID == 0 {
			l.Debug().Uint("user_id", userID).Msg("Invalid user ID in token claims")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("{\"error\": \"Unauthorized: invalid token\"}"))
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, ContextKeyUsername, username)
		ctx = context.WithValue(ctx, ContextKeyRole, role)
		ctx = context.WithValue(ctx, ContextKeyUserID, userID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)

	})
}
func ProtectAdminRoute(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(ContextKeyRole).(string)
		l := zerolog.Ctx(r.Context())
		l.Trace().Msg("Inside ProtectAdminRoute middleware")
		l.Debug().Str("role", role).Msg("Checking user role for admin access")
		if !ok || role != model.UserTypeAdmin {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("{\"error\": \"Forbidden: admin access required\"}"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func ProtectStaffRoute(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(ContextKeyRole).(string)
		l := zerolog.Ctx(r.Context())
		l.Trace().Msg("Inside ProtectStaffRoute middleware")
		l.Debug().Str("role", role).Msg("Checking user role for staff access")
		if !ok || (role != model.UserTypeStaff && role != model.UserTypeAdmin) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("{\"error\": \"Forbidden: staff access required\"}"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func ProtectOwnerRoute(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(ContextKeyRole).(string)
		l := zerolog.Ctx(r.Context())
		l.Trace().Msg("Inside ProtectOwnerRoute middleware")
		l.Debug().Str("role", role).Msg("Checking user role for owner access")
		if !ok || (role != model.UserTypeOwner && role != model.UserTypeStaff && role != model.UserTypeAdmin) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("{\"error\": \"Forbidden: owner access required\"}"))
			return
		}
		next.ServeHTTP(w, r)
	})
}
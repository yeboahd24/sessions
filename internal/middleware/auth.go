package middleware

import (
	"context"
	"net/http"
	"os"

	"music-session-app/internal/auth"

	"github.com/alexedwards/scs/v2"
)

type contextKey string

const ClaimsKey contextKey = "claims"

func RequireAuth(session *scs.SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// pull token from session
			token := session.GetString(r.Context(), "jwt")
			if token == "" {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			// parse token
			claims, err := auth.ParseToken(token, os.Getenv("JWT_SECRET"))
			if err != nil {
				session.Destroy(r.Context())
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			// set claims in context
			ctx := context.WithValue(r.Context(), ClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

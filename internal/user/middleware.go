package user

import (
	"context"
	"net/http"
)

type contextKey string

const authenticatedUserKey contextKey = "authenticated_user"

func (r *Repository) Authenticate(
	ctx context.Context,
	token string,
) (User, error) {
	if token == "" {
		return User{}, ErrNotFound
	}

	return r.GetUserByToken(ctx, hashToken(token))
}

func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r)

		if token == "" {
			http.Error(
				w,
				`{"error":"missing or invalid authorization token"}`,
				http.StatusUnauthorized,
			)
			return
		}

		u, err := h.service.repo.Authenticate(r.Context(), token)
		if err != nil {
			http.Error(
				w,
				`{"error":"unauthorized"}`,
				http.StatusUnauthorized,
			)
			return
		}

		ctx := context.WithValue(r.Context(), authenticatedUserKey, u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserFromContext(ctx context.Context) (User, bool) {
	u, ok := ctx.Value(authenticatedUserKey).(User)
	return u, ok
}

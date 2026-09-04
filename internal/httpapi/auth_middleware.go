package httpapi

import (
	"errors"
	"net/http"

	"flowerpress/internal/service"
)

func (s *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(sessionCookieName)
			if err != nil {
				if errors.Is(err, http.ErrNoCookie) {
					writeJSON(
						w,
						http.StatusUnauthorized,
						map[string]string{
							"error": "authentication required",
						},
					)

					return
				}

				writeJSON(
					w,
					http.StatusBadRequest,
					map[string]string{
						"error": "invalid session cookie",
					},
				)

				return
			}

			user, err := s.sessions.Validate(
				r.Context(),
				cookie.Value,
			)

			if err != nil {
				if errors.Is(err, service.ErrInvalidSession) {
					clearSessionCookie(w)

					writeJSON(
						w,
						http.StatusUnauthorized,
						map[string]string{
							"error": "authentication required",
						},
					)

					return
				}

				writeJSON(
					w,
					http.StatusInternalServerError,
					map[string]string{
						"error": "internal server error",
					},
				)

				return
			}

			ctx := withUser(
				r.Context(),
				user,
			)

			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			)
		},
	)
}

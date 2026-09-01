package httpapi

import (
	"errors"
	"net/http"

	"flowerpress/internal/service"
)

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type userResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var request registerRequest

	if !decodeJSON(w, r, &request) {
		return
	}

	user, err := s.users.Register(
		r.Context(),
		request.Username,
		request.Password,
	)

	switch {
	case errors.Is(err, service.ErrUsernameRequired):
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "username is required",
			},
		)
		return

	case errors.Is(err, service.ErrPasswordTooShort):
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "password must be at least 8 characters",
			},
		)
		return

	case errors.Is(err, service.ErrUsernameTaken):
		writeJSON(
			w,
			http.StatusConflict,
			map[string]string{
				"error": "username already taken",
			},
		)
		return

	case err != nil:
		writeJSON(
			w,
			http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
		return
	}

	writeJSON(
		w,
		http.StatusCreated,
		userResponse{
			ID:       user.ID,
			Username: user.Username,
		},
	)
}

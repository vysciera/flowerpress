package httpapi

import (
	"encoding/json"
	"net/http"
)

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "invalid body request",
			},
		)

		return false
	}

	return true
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(value)
}

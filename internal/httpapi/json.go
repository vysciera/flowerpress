package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

const maxRequestBodySize = 1 << 20 // 1 MiB - if even needed

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		maxRequestBodySize,
	)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		var maxBytesError *http.MaxBytesError

		if errors.As(err, &maxBytesError) {
			writeJSON(
				w,
				http.StatusRequestEntityTooLarge,
				map[string]string{
					"error": "request body too large",
				},
			)

			return false
		}

		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "invalid request body",
			},
		)

		return false
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "request body must contain a single JSON object",
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

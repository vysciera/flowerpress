package httpapi

import (
	"net/http"
	"testing"
	"time"
)

func TestNewHTTPServer(t *testing.T) {
	handler := http.NewServeMux()

	server := NewHTTPServer(
		":8080",
		handler,
	)

	if server.Addr != ":8080" {
		t.Fatalf(
			"expected address %q, got %q",
			":8080",
			server.Addr,
		)
	}

	if server.Handler != handler {
		t.Fatal("expected handler to be assigned")
	}

	// Look into Testify
	if server.ReadHeaderTimeout != 5*time.Second {
		t.Fatalf(
			"expected ReadHeaderTimeout %v, got %v",
			5*time.Second,
			server.ReadHeaderTimeout,
		)
	}

	if server.ReadTimeout != 15*time.Second {
		t.Fatalf(
			"expected ReadTimeout %v, got %v",
			15*time.Second,
			server.ReadTimeout,
		)
	}

	if server.WriteTimeout != 30*time.Second {
		t.Fatalf(
			"expected WriteTimeout %v, got %v",
			30*time.Second,
			server.WriteTimeout,
		)
	}

	if server.IdleTimeout != 60*time.Second {
		t.Fatalf(
			"expected IdleTimeout %v, got %v",
			60*time.Second,
			server.IdleTimeout,
		)
	}
}

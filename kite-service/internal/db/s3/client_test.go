package s3

import (
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/config"
)

// New must not contact the S3 server. Requiring a reachable server here made
// Kite refuse to start for self-hosted instances that don't use S3 at all.
func TestNewDoesNotRequireReachableServer(t *testing.T) {
	client, err := New(config.S3Config{Endpoint: "localhost:9000"})
	if err != nil {
		t.Fatalf("expected to build a client without reaching the server, got: %v", err)
	}
	if client == nil {
		t.Fatal("expected a client")
	}
}

package util

import (
	"strings"
	"testing"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/stretchr/testify/assert"
)

func TestAuditLogReason(t *testing.T) {
	assert.Equal(t, api.AuditLogReason(""), AuditLogReason(""))
	assert.Equal(t, api.AuditLogReason("Closed%20by%20M%C3%BCller%0Aspam%2Bscam"), AuditLogReason("Closed by Müller\nspam+scam"))
	assert.Equal(t, AuditLogReason(strings.Repeat("ü", 512)), AuditLogReason(strings.Repeat("ü", 600)))
}

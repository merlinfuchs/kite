package util

import (
	"errors"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/utils/httputil"
)

// maxAuditLogReasonLength is the most characters Discord allows in a reason.
const maxAuditLogReasonLength = 512

// AuditLogReason makes a reason fit the X-Audit-Log-Reason header. Discord
// rejects reasons over 512 characters and expects URL-encoded UTF-8, and a raw
// newline would make the request fail before it's sent.
func AuditLogReason(reason string) api.AuditLogReason {
	if utf8.RuneCountInString(reason) > maxAuditLogReasonLength {
		reason = string([]rune(reason)[:maxAuditLogReasonLength])
	}
	// Discord decodes "+" as a plus, so spaces must be %20.
	return api.AuditLogReason(strings.ReplaceAll(url.QueryEscape(reason), "+", "%20"))
}

func IsDiscordRestErrorCode(err error, code httputil.ErrorCode) bool {
	if err == nil {
		return false
	}

	var restErr *httputil.HTTPError
	if errors.As(err, &restErr) {
		return restErr.Code == code
	}

	return false
}

func IsDiscordRestStatusCode(err error, code int) bool {
	if err == nil {
		return false
	}

	var restErr *httputil.HTTPError
	if errors.As(err, &restErr) {
		return restErr.Status == code
	}

	return false
}

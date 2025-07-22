package util

import (
	"errors"

	"github.com/diamondburned/arikawa/v3/utils/httputil"
)

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

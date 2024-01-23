package fail

type PluginErrorCode int

const (
	PluginErrorCodeUnknown PluginErrorCode = 0
)

type PluginError struct {
	Code    PluginErrorCode `json:"code"`
	Message string          `json:"message"`
}

func (e *PluginError) Error() string {
	// TODO: figure out why this is needed
	if e == nil {
		return "err was nil (fix this!)"
	}
	return e.Message
}

func IsPluginErrorCode(err error, code PluginErrorCode) bool {
	if err == nil {
		return false
	}

	pluginError, ok := err.(*PluginError)
	if !ok {
		return false
	}

	return pluginError.Code == code
}

package fail

type ModuleErrorCode int

const (
	ModuleErrorCodeUnknown ModuleErrorCode = 0
)

type ModuleError struct {
	Code    ModuleErrorCode `json:"code"`
	Message string          `json:"message"`
}

func (e *ModuleError) Error() string {
	// TODO: figure out why this is needed
	if e == nil {
		return "err was nil (fix this!)"
	}
	return e.Message
}

func IsPluginErrorCode(err error, code ModuleErrorCode) bool {
	if err == nil {
		return false
	}

	pluginError, ok := err.(*ModuleError)
	if !ok {
		return false
	}

	return pluginError.Code == code
}

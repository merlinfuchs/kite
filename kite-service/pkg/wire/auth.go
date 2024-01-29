package wire

type AuthLoginStartRequest struct{}

type AuthCLIStartResponseData struct {
	Code string `json:"code"`
}

type AuthCLIStartResponse APIResponse[AuthCLIStartResponseData]

type AuthCLICallbackResponseData struct {
	Message string `json:"message"`
}

type AuthCLICallbackResponse APIResponse[AuthCLICallbackResponseData]

type AuthCLICheckResponseData struct {
	Pending bool   `json:"pending"`
	Token   string `json:"token,omitempty"`
}

type AuthCLICheckResponse APIResponse[AuthCLICheckResponseData]

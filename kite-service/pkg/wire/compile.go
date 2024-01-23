package wire

type CompileJSRequest struct {
	Source string `json:"source"`
}

type CompileJSResponseData struct {
	WASMBytes string `json:"wasm_bytes"`
}

type CompileJSResponse APIResponse[CompileJSResponseData]

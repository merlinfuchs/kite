package wire

type CompileRequest struct {
	Type   string `json:"type"`
	Source string `json:"source"`
}

type CompileResponseData struct {
	WASMBytes Base64 `json:"wasm_bytes"`
}

type CompileResponse APIResponse[CompileResponseData]

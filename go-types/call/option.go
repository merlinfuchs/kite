package call

import "time"

type CallConfig struct {
	Reason  string `json:"reason,omitempty"`
	Timeout int    `json:"timeout,omitempty"`
	Wait    bool   `json:"wait,omitempty"`
}

func NewCallConfig(opts ...CallOption) CallConfig {
	c := CallConfig{}
	for _, opt := range opts {
		opt(&c)
	}
	return c
}

type CallOption func(c *CallConfig)

// WithReason defines the reason for the call. Right now this is only used as the audit log reason for some Discord calls.
func WithReason(reason string) CallOption {
	return func(c *CallConfig) {
		c.Reason = reason
	}
}

// WithTimeout defines the timeout for the call.
func WithTimeout(timeout time.Duration) CallOption {
	return func(c *CallConfig) {
		c.Timeout = int(timeout.Milliseconds())
	}
}

// WithWait defines whether the call should wait for a response or return immediately (default: false).
func WithWait(wait bool) CallOption {
	return func(c *CallConfig) {
		c.Wait = wait
	}
}

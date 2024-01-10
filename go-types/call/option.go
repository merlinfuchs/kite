package call

import "time"

type CallConfig struct {
	Reason  string `json:"reason,omitempty"`
	Timeout int    `json:"timeout,omitempty"`
}

func NewCallConfig(opts ...CallOption) CallConfig {
	c := CallConfig{}
	for _, opt := range opts {
		opt(&c)
	}
	return c
}

type CallOption func(c *CallConfig)

func WithReason(reason string) CallOption {
	return func(c *CallConfig) {
		c.Reason = reason
	}
}

func WithTimeout(timeout time.Duration) CallOption {
	return func(c *CallConfig) {
		c.Timeout = int(timeout.Milliseconds())
	}
}

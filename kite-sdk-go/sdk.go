package kite

import (
	"time"

	"github.com/merlinfuchs/kite/kite-sdk-go/call"
	"github.com/merlinfuchs/kite/kite-sdk-go/internal"
)

func Sleep(duration time.Duration) error {
	_, err := internal.MakeCall(call.Call{
		Type: call.Sleep,
		Data: call.SleepCall{
			Duration: int(duration.Milliseconds()),
		},
	})
	if err != nil {
		return err
	}
	return nil
}

package internal

import (
	"encoding/json"
	"fmt"
	"unsafe"

	"github.com/merlinfuchs/kite/go-sdk/internal/util"
	"github.com/merlinfuchs/kite/go-types/call"
)

func MakeCall(c call.Call) (*call.CallResponse, error) {
	raw, err := json.Marshal(c)

	length := kiteCall(uint32(uintptr(unsafe.Pointer(&raw[0]))), uint32(len(raw)))
	if length == 0 {
		return nil, fmt.Errorf("no call response")
	}

	buf := make([]byte, length)
	ok := kiteGetCallResponse(uint32(uintptr(unsafe.Pointer(&buf[0]))))
	if ok != 0 {
		return nil, fmt.Errorf("failed to get call response")
	}

	var res call.CallResponse
	err = json.Unmarshal(buf, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal call response: %w", err)
	}

	if !res.Success {
		return nil, res.Error
	}

	return &res, nil
}

func CallWithResponse[R any](t call.CallType, opts []call.CallOption, d interface{}) (R, error) {
	res, err := MakeCall(call.Call{
		Type:   t,
		Config: call.NewCallConfig(opts...),
		Data:   d,
	})
	if err != nil {
		// TODO: better way to do this?
		r := new(R)
		return *r, err
	}

	return util.DecodeT[R](res.Data.(json.RawMessage))
}

func CallWithoutResponse(t call.CallType, opts []call.CallOption, d interface{}) error {
	_, err := MakeCall(call.Call{
		Type:   t,
		Config: call.NewCallConfig(opts...),
		Data:   d,
	})
	if err != nil {
		return err
	}

	return nil
}

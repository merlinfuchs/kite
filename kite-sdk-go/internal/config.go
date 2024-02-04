package internal

import (
	"encoding/json"
	"fmt"
	"unsafe"
)

func GetConfig() (map[string]interface{}, error) {
	size := kiteGetConfigSize()
	if size == 0 {
		return nil, nil
	}

	buf := make([]byte, size)

	offset := uint32(uintptr(unsafe.Pointer(&buf[0])))

	ok := kiteGetConfig(offset)
	if ok != 0 {
		return nil, fmt.Errorf("failed to get config")
	}

	var config map[string]interface{}
	err := json.Unmarshal(buf, &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

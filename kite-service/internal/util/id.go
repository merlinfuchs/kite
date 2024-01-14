package util

import "github.com/oklog/ulid/v2"

func UniqueID() string {
	return ulid.Make().String()
}

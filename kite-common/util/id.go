package util

import (
	gonanoid "github.com/matoous/go-nanoid/v2"
)

var alphabet = "0123456789abcdefghijklmnopqrstuvwxyz"

func UniqueID() string {
	id, _ := gonanoid.Generate(alphabet, 16)
	return id
}

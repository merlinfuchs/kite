//go:build embedapp
// +build embedapp

package kiteapp

import "embed"

//go:embed out/*
var OutFS embed.FS

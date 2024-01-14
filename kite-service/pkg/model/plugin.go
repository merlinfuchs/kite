package model

import "time"

type Plugin struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type PluginVersion struct {
	ID                    string
	PluginID              string
	VersionMajor          int32
	VersionMinor          int32
	VersionPatch          int32
	WasmBytes             []byte
	ManifestDefaultConfig map[string]string
	ManifestEvents        []string
	ManifestCommands      []string
	CreatedAt             time.Time
}

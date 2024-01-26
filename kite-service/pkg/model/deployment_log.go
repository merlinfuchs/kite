package model

import "time"

type DeploymentLogEntry struct {
	ID           uint64
	DeploymentID string
	Level        string
	Message      string
	CreatedAt    time.Time
}

type DeploymentLogSummary struct {
	DeploymentID string
	TotalCount   int
	ErrorCount   int
	WarnCount    int
	InfoCount    int
	DebugCount   int
}

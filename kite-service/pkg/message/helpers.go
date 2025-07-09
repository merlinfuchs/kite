package message

import (
	"fmt"
	"strconv"
	"strings"
)

func CustomIDModalResumePoint(resumePointID string) string {
	return fmt.Sprintf("resume:%s", resumePointID)
}

func DecodeCustomIDModalResumePoint(customID string) (string, bool) {
	if !strings.HasPrefix(customID, "resume:") {
		return "", false
	}

	return customID[len("resume:"):], true
}

func CustomIDMessageComponentResumePoint(resumePointID string, componentID int) string {
	return fmt.Sprintf("resume:%s_%d", resumePointID, componentID)
}

func DecodeCustomIDMessageComponentResumePoint(customID string) (string, int, bool) {
	if !strings.HasPrefix(customID, "resume:") {
		return "", 0, false
	}

	value := customID[len("resume:"):]

	parts := strings.Split(value, "_")
	if len(parts) != 2 {
		return "", 0, false
	}

	componentID, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, false
	}

	return parts[0], componentID, true
}

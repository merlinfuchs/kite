package message

import (
	"fmt"
	"strconv"
	"strings"
)

// ResumeCustomIDPrefix marks a component custom ID as addressing a resume
// point. The engine routes any interaction whose custom ID starts with it
// straight to a resume point lookup, so the prefix is reserved: a component's
// own custom ID comes from its flow source ID, which is authored by the tenant,
// and must never be able to land in this namespace.
const ResumeCustomIDPrefix = "resume:"

// IsReservedCustomID reports whether a custom ID would be read as addressing a
// resume point rather than the component it is attached to.
func IsReservedCustomID(customID string) bool {
	_, ok := strings.CutPrefix(customID, ResumeCustomIDPrefix)
	return ok
}

func CustomIDModalResumePoint(resumePointID string) string {
	return ResumeCustomIDPrefix + resumePointID
}

func DecodeCustomIDModalResumePoint(customID string) (string, bool) {
	return strings.CutPrefix(customID, ResumeCustomIDPrefix)
}

func CustomIDMessageComponentResumePoint(resumePointID string, componentID int) string {
	return fmt.Sprintf("%s%s_%d", ResumeCustomIDPrefix, resumePointID, componentID)
}

func DecodeCustomIDMessageComponentResumePoint(customID string) (string, int, bool) {
	value, ok := strings.CutPrefix(customID, ResumeCustomIDPrefix)
	if !ok {
		return "", 0, false
	}

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

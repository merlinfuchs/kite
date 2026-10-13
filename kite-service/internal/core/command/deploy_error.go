package command

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/utils/httputil"
	"github.com/kitecloud/kite/kite-service/internal/model"
)

// DeployError is a problem with an app's commands that its user has to fix.
type DeployError = model.CommandsDeployError

// isRejection reports whether Discord rejected the commands, rather than
// failing to process them. Nirn, the proxy in front of Discord, returns 408
// when Discord doesn't answer in time.
func isRejection(status int) bool {
	return status >= 400 && status < 500 &&
		status != http.StatusTooManyRequests &&
		status != httputil.StatusNirnTimeout
}

// discordDeployError turns Discord rejecting the commands into a message the
// user can act on.
func discordDeployError(httpErr *httputil.HTTPError, cmdData []api.CreateCommandData) *DeployError {
	if httpErr.Status == http.StatusUnauthorized {
		return &DeployError{Message: "Discord didn't accept the bot token. Update it in the settings."}
	}

	if httpErr.Errors != nil {
		if err := formError(httpErr.Errors, cmdData); err != nil {
			return err
		}
	}

	if httpErr.Message != "" {
		return &DeployError{Message: httpErr.Message}
	}
	return &DeployError{Message: fmt.Sprintf("Discord rejected the commands with status %d.", httpErr.Status)}
}

// formError describes the first of Discord's errors for an invalid form body.
// Discord keys the errors by the path of the invalid field, starting with the
// index of the command, like {"0": {"options": {"1": {"description": ...}}}}.
func formError(rawErrors []byte, cmdData []api.CreateCommandData) *DeployError {
	var errs map[string]any
	if err := json.Unmarshal(rawErrors, &errs); err != nil {
		return nil
	}

	path, message := firstFormError(errs, nil)
	if message == "" {
		return nil
	}
	if len(path) == 0 {
		return &DeployError{Message: message}
	}

	i, err := strconv.Atoi(path[0])
	if err != nil || i < 0 || i >= len(cmdData) {
		return &DeployError{Message: message}
	}

	// Walk the same path through the command to name the options on the way.
	raw, err := json.Marshal(cmdData[i])
	if err != nil {
		return &DeployError{Message: message, Command: cmdData[i].Name}
	}
	var node any
	if err := json.Unmarshal(raw, &node); err != nil {
		return &DeployError{Message: message, Command: cmdData[i].Name}
	}

	names := []string{"/" + cmdData[i].Name}
	field := ""
	for _, key := range path[1:] {
		switch v := node.(type) {
		case map[string]any:
			node = v[key]
			field = key
		case []any:
			j, err := strconv.Atoi(key)
			if err != nil || j < 0 || j >= len(v) {
				node = nil
				continue
			}
			node = v[j]
			// Only options are named in the location, not choices.
			if item, ok := node.(map[string]any); ok && field == "options" {
				if name, ok := item["name"].(string); ok {
					names = append(names, name)
				}
			}
			field = ""
		}
	}

	location := strings.Join(names, " ")
	if field != "" {
		field = strings.ReplaceAll(field, "_", " ")
		location = fmt.Sprintf("%s%s of %s", strings.ToUpper(field[:1]), field[1:], location)
	}

	return &DeployError{
		Message: fmt.Sprintf("%s: %s", location, message),
		Command: cmdData[i].Name,
	}
}

func firstFormError(errs map[string]any, path []string) ([]string, string) {
	if list, ok := errs["_errors"].([]any); ok && len(list) > 0 {
		if e, ok := list[0].(map[string]any); ok {
			if message, ok := e["message"].(string); ok {
				return path, message
			}
		}
	}

	keys := make([]string, 0, len(errs))
	for key := range errs {
		if key != "_errors" {
			keys = append(keys, key)
		}
	}
	// Map order is random, so sort to report the same error every time.
	slices.Sort(keys)

	for _, key := range keys {
		child, ok := errs[key].(map[string]any)
		if !ok {
			continue
		}
		if p, message := firstFormError(child, append(slices.Clone(path), key)); message != "" {
			return p, message
		}
	}

	return nil, ""
}

package flow

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/diamondburned/arikawa/v3/utils/sendpart"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

// Discord allows up to 10 files per message.
const maxMessageFiles = 10

var templatePlaceholderRe = regexp.MustCompile(`\{\{(.+?)\}\}`)

// evalTemplateFiles takes the placeholders that evaluate to a file, like the
// result of the channel transcript block, out of a template. It returns the
// template without them and the files, so they can be attached to a message.
func evalTemplateFiles(ctx context.Context, template string, evalCtx eval.Context) (string, []thing.FileValue) {
	if !strings.Contains(template, "{{") {
		return template, nil
	}

	var files []thing.FileValue
	res := templatePlaceholderRe.ReplaceAllStringFunc(template, func(match string) string {
		expression := strings.TrimSpace(match[2 : len(match)-2])

		// Only results and temporary variables can hold files. Other
		// placeholders are left alone so they aren't evaluated twice.
		if !strings.Contains(expression, "result") && !strings.Contains(expression, "var(") {
			return match
		}

		// Errors are left to the evaluation of the whole template.
		v, err := eval.Eval(ctx, expression, evalCtx)
		if err != nil || v.Type != thing.TypeFile {
			return match
		}

		files = append(files, v.File())
		return ""
	})

	return res, files
}

// toSendFiles turns files into files of a Discord request.
func toSendFiles(files []thing.FileValue) ([]sendpart.File, error) {
	if len(files) == 0 {
		return nil, nil
	}
	if len(files) > maxMessageFiles {
		return nil, fmt.Errorf("a message can have up to %d files, got %d", maxMessageFiles, len(files))
	}

	res := make([]sendpart.File, len(files))
	for i, file := range files {
		if len(file.Data) == 0 {
			return nil, fmt.Errorf("file %q is empty", file.Name)
		}
		res[i] = sendpart.File{
			Name:   file.Name,
			Reader: bytes.NewReader(file.Data),
		}
	}
	return res, nil
}

package flow

import (
	"fmt"
	"maps"
	"net/url"
	"regexp"
	"slices"
	"strings"

	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

var secretReferenceRe = regexp.MustCompile(`\bsecrets\.([A-Za-z0-9_]+)`)

// requestSecrets are the app's secrets a request block references, like
// {{secrets.API_KEY}}. Only the settings of requests can use secrets, so they
// can't end up in messages or logs by accident.
type requestSecrets struct {
	values  map[string]string
	evalCtx eval.Context
}

// newRequestSecrets fetches the secrets the given templates reference. Secrets
// are only decrypted when a flow uses them.
func newRequestSecrets(ctx *FlowContext, templates ...string) (*requestSecrets, error) {
	var names []string
	for _, template := range templates {
		for _, m := range secretReferenceRe.FindAllStringSubmatch(template, -1) {
			if !slices.Contains(names, m[1]) {
				names = append(names, m[1])
			}
		}
	}

	res := &requestSecrets{evalCtx: ctx.EvalCtx}
	if len(names) == 0 {
		return res, nil
	}
	if ctx.Secret == nil {
		return nil, fmt.Errorf("secrets can't be used here")
	}

	values, err := ctx.Secret.Secrets(ctx, names)
	if err != nil {
		return nil, err
	}
	env := make(map[string]any, len(values))
	for _, name := range names {
		value, ok := values[name]
		if !ok {
			return nil, fmt.Errorf("the app has no secret named %s", name)
		}
		env[name] = value
	}

	res.values = values
	res.evalCtx = eval.Context{
		Env:      maps.Clone(ctx.EvalCtx.Env),
		Patchers: ctx.EvalCtx.Patchers,
	}
	if res.evalCtx.Env == nil {
		res.evalCtx.Env = eval.Env{}
	}
	res.evalCtx.Env["secrets"] = env
	return res, nil
}

// EvalTemplate evaluates a template of the request, where the secrets are
// available.
func (s *requestSecrets) EvalTemplate(ctx *FlowContext, template string) (thing.Thing, error) {
	res, err := eval.EvalTemplate(ctx, template, s.evalCtx)
	if err != nil {
		return thing.Null, s.Redact(fmt.Errorf("failed to evaluate template: %w", err))
	}
	return res, nil
}

// Redact removes the values of the secrets from an error, which can contain
// them, e.g. a failed request's error contains its URL.
func (s *requestSecrets) Redact(err error) error {
	if err == nil || len(s.values) == 0 {
		return err
	}

	msg := err.Error()
	for _, value := range s.values {
		if value == "" {
			continue
		}
		for _, form := range []string{value, url.QueryEscape(value), url.PathEscape(value)} {
			msg = strings.ReplaceAll(msg, form, "[secret]")
		}
	}
	if msg == err.Error() {
		return err
	}
	return redactedError(msg)
}

type redactedError string

func (e redactedError) Error() string {
	return string(e)
}

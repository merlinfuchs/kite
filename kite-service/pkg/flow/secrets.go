package flow

import (
	"errors"
	"fmt"
	"maps"
	"net/url"
	"regexp"
	"slices"
	"strings"

	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

var (
	placeholderRe = regexp.MustCompile(`(?s)\{\{(.*?)\}\}`)
	// Not after a dot, as webhook.data.secrets.count isn't a secret.
	secretReferenceRe = regexp.MustCompile(`(?:^|[^.\w])secrets\.([A-Za-z_][A-Za-z0-9_]*)`)
	plainSecretRe     = regexp.MustCompile(`^\s*secrets\.[A-Za-z_][A-Za-z0-9_]*\s*$`)
	// Any use of the secrets, also like secrets?.NAME, which can read the
	// ones another placeholder of the template references.
	secretMentionRe = regexp.MustCompile(`(?:^|[^.\w])secrets\b`)
)

// secretNames are the secrets the templates reference. Only placeholders are
// searched, as text like secrets.txt in a URL doesn't reference a secret.
func secretNames(templates []string) []string {
	var names []string
	for _, template := range templates {
		for _, p := range placeholderRe.FindAllStringSubmatch(template, -1) {
			for _, m := range secretReferenceRe.FindAllStringSubmatch(p[1], -1) {
				if !slices.Contains(names, m[1]) {
					names = append(names, m[1])
				}
			}
		}
	}
	return names
}

// onlyPlainSecrets reports whether every placeholder of template that
// references a secret is just the secret, like {{secrets.TOKEN}}. An
// expression could compare a part of it instead.
func onlyPlainSecrets(template string) bool {
	for _, p := range placeholderRe.FindAllStringSubmatch(template, -1) {
		if secretMentionRe.MatchString(p[1]) && !plainSecretRe.MatchString(p[1]) {
			return false
		}
	}
	return true
}

// flowSecrets are the app's secrets the templates of a setting reference,
// like {{secrets.API_KEY}}. Only settings evaluated with them can use secrets,
// like those of requests, so they can't end up in messages or logs by
// accident.
type flowSecrets struct {
	values  map[string]string
	evalCtx eval.Context
}

// newPlainFlowSecrets is newFlowSecrets for a template whose secrets can only
// be used as they are, not in an expression that could use a part of one.
func newPlainFlowSecrets(ctx *FlowContext, template string) (*flowSecrets, error) {
	if !onlyPlainSecrets(template) {
		return nil, fmt.Errorf("secrets can only be used as {{secrets.NAME}}, not in an expression")
	}
	return newFlowSecrets(ctx, template)
}

// newFlowSecrets fetches the secrets the given templates reference. Secrets
// are only decrypted when a flow uses them.
func newFlowSecrets(ctx *FlowContext, templates ...string) (*flowSecrets, error) {
	names := secretNames(templates)
	res := &flowSecrets{evalCtx: ctx.EvalCtx}
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

// EvalTemplate evaluates a template of the setting, where the secrets are
// available.
func (s *flowSecrets) EvalTemplate(ctx *FlowContext, template string) (thing.Thing, error) {
	res, err := eval.EvalTemplate(ctx, template, s.evalCtx)
	if err != nil {
		return thing.Null, s.Redact(fmt.Errorf("failed to evaluate template: %w", err))
	}
	return res, nil
}

// EvalTemplateRaw is like FlowContext.EvalTemplateRaw, for a setting that
// has to be exactly the secret, like a signing key.
func (s *flowSecrets) EvalTemplateRaw(ctx *FlowContext, template string) (thing.Thing, error) {
	evalCtx := s.evalCtx
	evalCtx.Raw = true
	res, err := eval.EvalTemplateKeepSpace(ctx, template, evalCtx)
	if err != nil {
		return thing.Null, s.Redact(fmt.Errorf("failed to evaluate template: %w", err))
	}
	return res, nil
}

// Redact removes the values of the secrets from an error, which can contain
// them, e.g. a failed request's error contains its URL.
func (s *flowSecrets) Redact(err error) error {
	if err == nil || len(s.values) == 0 {
		return err
	}

	msg := err.Error()
	// A failed request's URL can contain a secret in any encoding, so only its
	// host is kept.
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		if u, parseErr := url.Parse(urlErr.URL); parseErr == nil {
			msg = strings.ReplaceAll(msg, urlErr.URL, u.Scheme+"://"+u.Host)
		}
	}
	msg = s.redactString(msg)
	if msg == err.Error() {
		return err
	}
	return redactedError(msg)
}

func (s *flowSecrets) redactString(msg string) string {
	for _, value := range s.values {
		// Replacing very short values would mangle the error and give away
		// the value.
		if len(value) < 4 {
			continue
		}
		for _, form := range []string{value, url.QueryEscape(value), url.PathEscape(value)} {
			msg = strings.ReplaceAll(msg, form, "[secret]")
		}
	}
	return msg
}

type redactedError string

func (e redactedError) Error() string {
	return string(e)
}

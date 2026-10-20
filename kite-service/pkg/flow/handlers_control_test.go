package flow

import (
	"context"
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/webhook"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// executeWebhookVerify runs a verify block with data for event, whose
// branches log "valid" and "invalid".
func executeWebhookVerify(t *testing.T, event *webhook.Event, data FlowNodeData, secrets *secretTestProvider) ([]string, error) {
	return executeWebhookVerifyWithDiscord(t, event, data, secrets, &provider.MockDiscordProvider{})
}

func executeWebhookVerifyWithDiscord(t *testing.T, event *webhook.Event, data FlowNodeData, secrets *secretTestProvider, discord provider.DiscordProvider) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	logs := &recordingLogProvider{}
	c := NewContext(
		ctx,
		5*time.Second,
		&webhookContextData{event: event},
		FlowProviders{Discord: discord, Log: logs, Secret: secrets},
		FlowContextLimits{MaxStackDepth: 10, MaxOperations: 1000, MaxCredits: 1000},
		eval.NewContext(eval.Env{"webhook": eval.NewWebhookEnv(event)}),
		nil,
	)
	t.Cleanup(c.Cancel)

	logNode := func(id string, message string) *CompiledFlowNode {
		return &CompiledFlowNode{ID: id, Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogLevel: provider.LogLevelInfo, LogMessage: message}}
	}
	node := &CompiledFlowNode{
		ID:   "1",
		Type: FlowNodeTypeControlWebhookVerify,
		Data: data,
		Children: ConnectedFlowNodes{Handles: map[string][]*CompiledFlowNode{
			"valid":   {logNode("2", "valid")},
			"invalid": {logNode("3", "invalid")},
		}},
	}
	err := node.Execute(c)
	return logs.messages, err
}

func githubSignature(key string, body string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(body))
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestWebhookVerifyPreset(t *testing.T) {
	secrets := &secretTestProvider{values: map[string]string{"GITHUB_SECRET": "s3cr3t"}}
	data := FlowNodeData{WebhookVerifyPreset: "github", WebhookVerifyKey: "{{secrets.GITHUB_SECRET}}"}

	event := &webhook.Event{Body: "{}", Headers: map[string]string{"x-hub-signature-256": githubSignature("s3cr3t", "{}")}}
	logs, err := executeWebhookVerify(t, event, data, secrets)
	require.NoError(t, err)
	assert.Equal(t, []string{"valid"}, logs)

	event = &webhook.Event{Body: "{}", Headers: map[string]string{"x-hub-signature-256": githubSignature("wrong", "{}")}}
	logs, err = executeWebhookVerify(t, event, data, secrets)
	require.NoError(t, err)
	assert.Equal(t, []string{"invalid"}, logs)
}

func TestWebhookVerifyCustom(t *testing.T) {
	// The newline at the end is part of what's signed.
	body := "{}\n"
	data := FlowNodeData{
		WebhookVerifyAlgorithm: webhook.AlgorithmHMACSHA256,
		WebhookVerifyKey:       "key",
		WebhookVerifyMessage:   "{{webhook.headers['x-id']}}.{{webhook.body}}",
		WebhookVerifySignature: "{{ trimPrefix(webhook.headers['x-signature'], 'sha256=') }}",
	}

	event := &webhook.Event{Body: body, Headers: map[string]string{
		"x-id":        "1",
		"x-signature": githubSignature("key", "1."+body),
	}}
	logs, err := executeWebhookVerify(t, event, data, &secretTestProvider{})
	require.NoError(t, err)
	assert.Equal(t, []string{"valid"}, logs)

	// A request without the headers is only invalid, even though the
	// signature template fails for it.
	logs, err = executeWebhookVerify(t, &webhook.Event{Body: body}, data, &secretTestProvider{})
	require.NoError(t, err)
	assert.Contains(t, logs, "invalid")
}

// Spaces around the body are part of what's signed too.
func TestWebhookVerifyCustomKeepsSpaces(t *testing.T) {
	data := FlowNodeData{
		WebhookVerifyKey:       "key",
		WebhookVerifyMessage:   "{{webhook.body}}\n",
		WebhookVerifySignature: "{{ trimPrefix(webhook.headers['x-signature'], 'sha256=') }}",
	}
	event := &webhook.Event{Body: "{}", Headers: map[string]string{"x-signature": githubSignature("key", "{}\n")}}
	logs, err := executeWebhookVerify(t, event, data, &secretTestProvider{})
	require.NoError(t, err)
	assert.Equal(t, []string{"valid"}, logs)
}

// Escapes like \n in a JSON body are part of what's signed, not newlines.
func TestWebhookVerifyCustomRawBody(t *testing.T) {
	body := `{"message":"a\nb"}`
	data := FlowNodeData{
		WebhookVerifyKey:       "key",
		WebhookVerifyMessage:   "{{webhook.body}}",
		WebhookVerifySignature: "{{ trimPrefix(webhook.headers['x-signature'], 'sha256=') }}",
	}
	event := &webhook.Event{Body: body, Headers: map[string]string{"x-signature": githubSignature("key", body)}}
	logs, err := executeWebhookVerify(t, event, data, &secretTestProvider{})
	require.NoError(t, err)
	assert.Equal(t, []string{"valid"}, logs)
}

// A template that doesn't compile is a mistake in the settings, not in the
// request.
func TestWebhookVerifyCustomInvalidTemplate(t *testing.T) {
	data := FlowNodeData{
		WebhookVerifyKey:       "key",
		WebhookVerifyMessage:   "{{webhook.body}}",
		WebhookVerifySignature: "{{ webhook.headers[ }}",
	}
	_, err := executeWebhookVerify(t, &webhook.Event{Body: "{}"}, data, &secretTestProvider{})
	assert.Error(t, err)
}

// ER:LC's key is built in, so a key left from another preset isn't fetched.
func TestWebhookVerifyERLCIgnoresKey(t *testing.T) {
	secrets := &secretTestProvider{values: map[string]string{"KEY": "s3cr3t"}}
	data := FlowNodeData{WebhookVerifyPreset: webhook.PresetERLC, WebhookVerifyKey: "{{secrets.KEY}}"}
	logs, err := executeWebhookVerify(t, &webhook.Event{Body: "{}"}, data, secrets)
	require.NoError(t, err)
	assert.Equal(t, []string{"invalid"}, logs)
	assert.Empty(t, secrets.asked)
}

func TestWebhookVerifySettingErrors(t *testing.T) {
	event := &webhook.Event{Body: "{}"}
	secrets := &secretTestProvider{values: map[string]string{"KEY": "s3cr3t"}}

	cases := map[string]FlowNodeData{
		"unknown preset":  {WebhookVerifyPreset: "unknown", WebhookVerifyKey: "key"},
		"empty key":       {WebhookVerifyPreset: "github"},
		"secret in expr":  {WebhookVerifyPreset: "github", WebhookVerifyKey: "{{secrets.KEY[0:1]}}"},
		"unknown secret":  {WebhookVerifyPreset: "github", WebhookVerifyKey: "{{secrets.MISSING}}"},
		"invalid ed25519": {WebhookVerifyAlgorithm: webhook.AlgorithmEd25519, WebhookVerifyKey: "nope"},
	}
	for name, data := range cases {
		_, err := executeWebhookVerify(t, event, data, secrets)
		assert.Error(t, err, name)
	}
}

func TestWebhookVerifyOnlyInWebhooks(t *testing.T) {
	err := (&CompiledFlowNode{ID: "1", Type: FlowNodeTypeControlWebhookVerify}).Execute(
		secretsTestContext(t, &secretTestProvider{}, nil, &provider.MockLogProvider{}),
	)
	assert.Error(t, err)
}

type verifyKeyDiscordProvider struct {
	provider.MockDiscordProvider
	key string
	err error
}

func (p *verifyKeyDiscordProvider) AppVerifyKey(ctx context.Context) (string, error) {
	return p.key, p.err
}

// The Discord preset uses the app's public key.
func TestWebhookVerifyDiscordAppKey(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	now := time.Now()
	ts := strconv.FormatInt(now.Unix(), 10)
	event := &webhook.Event{Body: "{}", ReceivedAt: now, Headers: map[string]string{
		"x-signature-timestamp": ts,
		"x-signature-ed25519":   hex.EncodeToString(ed25519.Sign(privateKey, []byte(ts+"{}"))),
	}}
	data := FlowNodeData{WebhookVerifyPreset: webhook.PresetDiscord}

	logs, err := executeWebhookVerifyWithDiscord(t, event, data, &secretTestProvider{}, &verifyKeyDiscordProvider{key: hex.EncodeToString(publicKey)})
	require.NoError(t, err)
	assert.Equal(t, []string{"valid"}, logs)

	_, err = executeWebhookVerifyWithDiscord(t, event, data, &secretTestProvider{}, &verifyKeyDiscordProvider{err: errors.New("not fetched yet")})
	assert.ErrorContains(t, err, "not fetched yet")
}

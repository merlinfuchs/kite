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

func TestControlLoopMissingEnd(t *testing.T) {
	entry, err := CompileCommand(FlowData{
		Nodes: []FlowNode{
			{ID: "entry", Type: FlowNodeTypeEntryCommand, Data: FlowNodeData{Name: "test", Description: "Test"}},
			{ID: "loop", Type: FlowNodeTypeControlLoop, Data: FlowNodeData{LoopCount: "3"}},
			{ID: "each", Type: FlowNodeTypeControlLoopEach},
			{ID: "log", Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogLevel: provider.LogLevelInfo, LogMessage: "iter"}},
		},
		Edges: []FlowEdge{
			{Source: "entry", Target: "loop"},
			{Source: "loop", Target: "each"},
			{Source: "each", Target: "log"},
		},
	})
	require.NoError(t, err)

	c, p := newSleepTestContext(t, &TestContextData{})
	require.NoError(t, entry.Execute(c))
	assert.Equal(t, []string{"iter", "iter", "iter"}, p.log.messages)
}

func TestControlLoopMissingEach(t *testing.T) {
	entry, err := CompileCommand(FlowData{
		Nodes: []FlowNode{
			{ID: "entry", Type: FlowNodeTypeEntryCommand, Data: FlowNodeData{Name: "test", Description: "Test"}},
			{ID: "loop", Type: FlowNodeTypeControlLoop, Data: FlowNodeData{LoopCount: "3"}},
			{ID: "end", Type: FlowNodeTypeControlLoopEnd},
			{ID: "log", Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogLevel: provider.LogLevelInfo, LogMessage: "done"}},
		},
		Edges: []FlowEdge{
			{Source: "entry", Target: "loop"},
			{Source: "loop", Target: "end"},
			{Source: "end", Target: "log"},
		},
	})
	require.NoError(t, err)

	c, p := newSleepTestContext(t, &TestContextData{})
	require.NoError(t, entry.Execute(c))
	assert.Equal(t, []string{"done"}, p.log.messages)
}

func TestControlLoopMissingBoth(t *testing.T) {
	entry, err := CompileCommand(FlowData{
		Nodes: []FlowNode{
			{ID: "entry", Type: FlowNodeTypeEntryCommand, Data: FlowNodeData{Name: "test", Description: "Test"}},
			{ID: "loop", Type: FlowNodeTypeControlLoop, Data: FlowNodeData{LoopCount: "3"}},
		},
		Edges: []FlowEdge{
			{Source: "entry", Target: "loop"},
		},
	})
	require.NoError(t, err)

	c, p := newSleepTestContext(t, &TestContextData{})
	require.NoError(t, entry.Execute(c))
	assert.Empty(t, p.log.messages)
}

func TestControlLoopExit(t *testing.T) {
	entry, err := CompileCommand(FlowData{
		Nodes: []FlowNode{
			{ID: "entry", Type: FlowNodeTypeEntryCommand, Data: FlowNodeData{Name: "test", Description: "Test"}},
			{ID: "loop", Type: FlowNodeTypeControlLoop, Data: FlowNodeData{LoopCount: "5"}},
			{ID: "each", Type: FlowNodeTypeControlLoopEach},
			{ID: "log_each", Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogLevel: provider.LogLevelInfo, LogMessage: "iter"}},
			{ID: "exit", Type: FlowNodeTypeControlLoopExit},
			{ID: "end", Type: FlowNodeTypeControlLoopEnd},
			{ID: "log_end", Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogLevel: provider.LogLevelInfo, LogMessage: "done"}},
		},
		Edges: []FlowEdge{
			{Source: "entry", Target: "loop"},
			{Source: "loop", Target: "each"},
			{Source: "each", Target: "log_each"},
			{Source: "log_each", Target: "exit"},
			{Source: "loop", Target: "end"},
			{Source: "end", Target: "log_end"},
		},
	})
	require.NoError(t, err)

	c, p := newSleepTestContext(t, &TestContextData{})
	require.NoError(t, entry.Execute(c))
	assert.Equal(t, []string{"iter", "done"}, p.log.messages)
}

// The flow ends in the first iteration, so neither the other iterations, the
// blocks after the loop, nor the entry's second branch run.
func TestControlFlowEnd(t *testing.T) {
	entry, err := CompileCommand(FlowData{
		Nodes: []FlowNode{
			{ID: "entry", Type: FlowNodeTypeEntryCommand, Data: FlowNodeData{Name: "test", Description: "Test"}},
			{ID: "loop", Type: FlowNodeTypeControlLoop, Data: FlowNodeData{LoopCount: "5"}},
			{ID: "each", Type: FlowNodeTypeControlLoopEach},
			{ID: "log_each", Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogLevel: provider.LogLevelInfo, LogMessage: "iter"}},
			{ID: "stop", Type: FlowNodeTypeControlFlowEnd},
			{ID: "end", Type: FlowNodeTypeControlLoopEnd},
			{ID: "log_end", Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogLevel: provider.LogLevelInfo, LogMessage: "done"}},
			{ID: "log_other", Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogLevel: provider.LogLevelInfo, LogMessage: "other"}},
		},
		Edges: []FlowEdge{
			{Source: "entry", Target: "loop"},
			{Source: "entry", Target: "log_other"},
			{Source: "loop", Target: "each"},
			{Source: "each", Target: "log_each"},
			{Source: "log_each", Target: "stop"},
			{Source: "loop", Target: "end"},
			{Source: "end", Target: "log_end"},
		},
	})
	require.NoError(t, err)

	c, p := newSleepTestContext(t, &TestContextData{})
	require.NoError(t, entry.Execute(c))
	assert.Equal(t, []string{"iter"}, p.log.messages)
}

func TestConditionMatchesInEveryLoopIteration(t *testing.T) {
	entry, err := CompileCommand(FlowData{
		Nodes: []FlowNode{
			{ID: "entry", Type: FlowNodeTypeEntryCommand, Data: FlowNodeData{Name: "test", Description: "Test"}},
			{ID: "loop", Type: FlowNodeTypeControlLoop, Data: FlowNodeData{LoopCount: "2"}},
			{ID: "each", Type: FlowNodeTypeControlLoopEach},
			{ID: "end", Type: FlowNodeTypeControlLoopEnd},
			{ID: "condition", Type: FlowNodeTypeControlConditionCompare, Data: FlowNodeData{ConditionBaseValue: "a"}},
			{ID: "item", Type: FlowNodeTypeControlConditionItemCompare, Data: FlowNodeData{ConditionItemMode: ComparsionModeEqual, ConditionItemValue: "a"}},
			{ID: "else", Type: FlowNodeTypeControlConditionItemElse},
			{ID: "match", Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogLevel: provider.LogLevelInfo, LogMessage: "match"}},
			{ID: "no-match", Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogLevel: provider.LogLevelInfo, LogMessage: "else"}},
		},
		Edges: []FlowEdge{
			{Source: "entry", Target: "loop"},
			{Source: "loop", Target: "each"},
			{Source: "loop", Target: "end"},
			{Source: "each", Target: "condition"},
			{Source: "condition", Target: "item"},
			{Source: "condition", Target: "else"},
			{Source: "item", Target: "match"},
			{Source: "else", Target: "no-match"},
		},
	})
	require.NoError(t, err)

	c, p := newSleepTestContext(t, &TestContextData{})
	require.NoError(t, entry.Execute(c))
	assert.Equal(t, []string{"match", "match"}, p.log.messages)
}

// The count is 1 in the first iteration and 2 in the second, so the branch
// matches once and else runs once.
func TestConditionRunsElseAfterAnEarlierIterationMatched(t *testing.T) {
	entry, err := CompileCommand(FlowData{
		Nodes: []FlowNode{
			{ID: "entry", Type: FlowNodeTypeEntryCommand, Data: FlowNodeData{Name: "test", Description: "Test"}},
			{ID: "loop", Type: FlowNodeTypeControlLoop, Data: FlowNodeData{LoopCount: "2"}},
			{ID: "each", Type: FlowNodeTypeControlLoopEach},
			{ID: "end", Type: FlowNodeTypeControlLoopEnd},
			{ID: "count", Type: FlowNodeTypeActionExpressionEvaluate, Data: FlowNodeData{Expression: "(result('count') ?? 0) + 1"}},
			{ID: "condition", Type: FlowNodeTypeControlConditionCompare, Data: FlowNodeData{ConditionBaseValue: "count {{result('count')}}"}},
			{ID: "item", Type: FlowNodeTypeControlConditionItemCompare, Data: FlowNodeData{ConditionItemMode: ComparsionModeEqual, ConditionItemValue: "count 1"}},
			{ID: "else", Type: FlowNodeTypeControlConditionItemElse},
			{ID: "match", Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogLevel: provider.LogLevelInfo, LogMessage: "match"}},
			{ID: "no-match", Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogLevel: provider.LogLevelInfo, LogMessage: "else"}},
		},
		Edges: []FlowEdge{
			{Source: "entry", Target: "loop"},
			{Source: "loop", Target: "each"},
			{Source: "loop", Target: "end"},
			{Source: "each", Target: "count"},
			{Source: "count", Target: "condition"},
			{Source: "condition", Target: "item"},
			{Source: "condition", Target: "else"},
			{Source: "item", Target: "match"},
			{Source: "else", Target: "no-match"},
		},
	})
	require.NoError(t, err)

	c, p := newSleepTestContext(t, &TestContextData{})
	require.NoError(t, entry.Execute(c))
	assert.Equal(t, []string{"match", "else"}, p.log.messages)
}

// The inner loop exits in its first iteration every time the outer loop runs
// it, so it logs once per outer iteration.
func TestLoopRunsAgainAfterExit(t *testing.T) {
	entry, err := CompileCommand(FlowData{
		Nodes: []FlowNode{
			{ID: "entry", Type: FlowNodeTypeEntryCommand, Data: FlowNodeData{Name: "test", Description: "Test"}},
			{ID: "condition", Type: FlowNodeTypeControlConditionCompare, Data: FlowNodeData{ConditionBaseValue: "a", ConditionAllowMultiple: true}},
			{ID: "item1", Type: FlowNodeTypeControlConditionItemCompare, Data: FlowNodeData{ConditionItemMode: ComparsionModeEqual, ConditionItemValue: "a"}},
			{ID: "item2", Type: FlowNodeTypeControlConditionItemCompare, Data: FlowNodeData{ConditionItemMode: ComparsionModeEqual, ConditionItemValue: "a"}},
			{ID: "loop", Type: FlowNodeTypeControlLoop, Data: FlowNodeData{LoopCount: "3"}},
			{ID: "each", Type: FlowNodeTypeControlLoopEach},
			{ID: "end", Type: FlowNodeTypeControlLoopEnd},
			{ID: "log", Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogLevel: provider.LogLevelInfo, LogMessage: "each"}},
			{ID: "exit", Type: FlowNodeTypeControlLoopExit},
		},
		Edges: []FlowEdge{
			{Source: "entry", Target: "condition"},
			{Source: "condition", Target: "item1"},
			{Source: "condition", Target: "item2"},
			// Both branches run the same loop.
			{Source: "item1", Target: "loop"},
			{Source: "item2", Target: "loop"},
			{Source: "loop", Target: "each"},
			{Source: "loop", Target: "end"},
			{Source: "each", Target: "log"},
			{Source: "log", Target: "exit"},
		},
	})
	require.NoError(t, err)

	c, p := newSleepTestContext(t, &TestContextData{})
	require.NoError(t, entry.Execute(c))
	assert.Equal(t, []string{"each", "each"}, p.log.messages)
}

// The matched branch runs the condition again through a cycle, where it
// doesn't match. The outer run still matched, so its else doesn't run.
func TestConditionKeepsItsMatchWhenRunInsideItsBranch(t *testing.T) {
	entry, err := CompileCommand(FlowData{
		Nodes: []FlowNode{
			{ID: "entry", Type: FlowNodeTypeEntryCommand, Data: FlowNodeData{Name: "test", Description: "Test"}},
			{ID: "count", Type: FlowNodeTypeActionExpressionEvaluate, Data: FlowNodeData{Expression: "(result('count') ?? 0) + 1"}},
			{ID: "condition", Type: FlowNodeTypeControlConditionCompare, Data: FlowNodeData{ConditionBaseValue: "count {{result('count')}}"}},
			{ID: "item", Type: FlowNodeTypeControlConditionItemCompare, Data: FlowNodeData{ConditionItemMode: ComparsionModeEqual, ConditionItemValue: "count 1"}},
			{ID: "else", Type: FlowNodeTypeControlConditionItemElse},
			{ID: "match", Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogLevel: provider.LogLevelInfo, LogMessage: "match"}},
			{ID: "no-match", Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogLevel: provider.LogLevelInfo, LogMessage: "else"}},
		},
		Edges: []FlowEdge{
			{Source: "entry", Target: "count"},
			{Source: "count", Target: "condition"},
			{Source: "condition", Target: "item"},
			{Source: "condition", Target: "else"},
			{Source: "item", Target: "match"},
			{Source: "match", Target: "count"},
			{Source: "else", Target: "no-match"},
		},
	})
	require.NoError(t, err)

	c, p := newSleepTestContext(t, &TestContextData{})
	require.NoError(t, entry.Execute(c))
	assert.Equal(t, []string{"match", "else"}, p.log.messages)
}

// The items of a condition don't count towards the operations limit, so a
// condition can have more items than the limit allows operations.
func TestConditionItemsAreOneOperation(t *testing.T) {
	data := FlowData{
		Nodes: []FlowNode{
			{ID: "entry", Type: FlowNodeTypeEntryCommand, Data: FlowNodeData{Name: "test", Description: "Test"}},
			{ID: "condition", Type: FlowNodeTypeControlConditionCompare, Data: FlowNodeData{ConditionBaseValue: "149"}},
		},
		Edges: []FlowEdge{{Source: "entry", Target: "condition"}},
	}
	for i := range 150 {
		value := strconv.Itoa(i)
		data.Nodes = append(data.Nodes,
			FlowNode{ID: "item-" + value, Type: FlowNodeTypeControlConditionItemCompare, Data: FlowNodeData{ConditionItemMode: ComparsionModeEqual, ConditionItemValue: value}},
			FlowNode{ID: "log-" + value, Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogLevel: provider.LogLevelInfo, LogMessage: value}},
		)
		data.Edges = append(data.Edges,
			FlowEdge{Source: "condition", Target: "item-" + value},
			FlowEdge{Source: "item-" + value, Target: "log-" + value},
		)
	}
	entry, err := CompileCommand(data)
	require.NoError(t, err)

	c, p := newSleepTestContext(t, &TestContextData{})
	c.MaxOperations = 10
	require.NoError(t, entry.Execute(c))
	assert.Equal(t, []string{"149"}, p.log.messages)
}

// A condition that runs itself again through its else still counts each run.
func TestConditionCycleReachesOperationsLimit(t *testing.T) {
	entry, err := CompileCommand(FlowData{
		Nodes: []FlowNode{
			{ID: "entry", Type: FlowNodeTypeEntryCommand, Data: FlowNodeData{Name: "test", Description: "Test"}},
			{ID: "condition", Type: FlowNodeTypeControlConditionCompare, Data: FlowNodeData{ConditionBaseValue: "a"}},
			{ID: "else", Type: FlowNodeTypeControlConditionItemElse},
		},
		Edges: []FlowEdge{
			{Source: "entry", Target: "condition"},
			{Source: "condition", Target: "else"},
			{Source: "else", Target: "condition"},
		},
	})
	require.NoError(t, err)

	c, _ := newSleepTestContext(t, &TestContextData{})
	c.MaxOperations = 10
	assert.ErrorContains(t, entry.Execute(c), "operations limit exceeded")
}

// Blocks connected to a condition directly, instead of through an item, still
// count as operations.
func TestConditionConnectedToItselfReachesOperationsLimit(t *testing.T) {
	entry, err := CompileCommand(FlowData{
		Nodes: []FlowNode{
			{ID: "entry", Type: FlowNodeTypeEntryCommand, Data: FlowNodeData{Name: "test", Description: "Test"}},
			{ID: "condition", Type: FlowNodeTypeControlConditionCompare, Data: FlowNodeData{ConditionBaseValue: "a"}},
		},
		Edges: []FlowEdge{
			{Source: "entry", Target: "condition"},
			{Source: "condition", Target: "condition"},
		},
	})
	require.NoError(t, err)

	c, _ := newSleepTestContext(t, &TestContextData{})
	c.MaxOperations = 10
	assert.ErrorContains(t, entry.Execute(c), "operations limit exceeded")
}

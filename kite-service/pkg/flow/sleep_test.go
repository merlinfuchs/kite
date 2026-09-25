package flow

import (
	"context"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

type recordingResumePointProvider struct {
	created []ResumePoint
}

func (p *recordingResumePointProvider) CreateResumePoint(ctx context.Context, rp ResumePoint) (ResumePoint, error) {
	p.created = append(p.created, rp)
	return rp, nil
}

type recordingLogProvider struct {
	messages []string
}

func (p *recordingLogProvider) CreateLogEntry(ctx context.Context, level provider.LogLevel, message string) {
	p.messages = append(p.messages, message)
}

type sleepTestProviders struct {
	discord     *TestDiscordProvider
	log         *recordingLogProvider
	resumePoint *recordingResumePointProvider
}

func newSleepTestContext(t *testing.T, data FlowContextData) (*FlowContext, sleepTestProviders) {
	t.Helper()

	p := sleepTestProviders{
		discord:     &TestDiscordProvider{},
		log:         &recordingLogProvider{},
		resumePoint: &recordingResumePointProvider{},
	}
	c := NewContext(
		context.Background(),
		5*time.Second,
		data,
		FlowProviders{Discord: p.discord, Log: p.log, ResumePoint: p.resumePoint},
		FlowContextLimits{MaxStackDepth: 20, MaxOperations: 1000, MaxCredits: 1000},
		eval.NewContext(eval.Env{}),
		nil,
	)
	t.Cleanup(c.Cancel)
	return c, p
}

// sleepFlow is a command that sleeps for seconds and then logs "after".
func sleepFlow(t *testing.T, seconds string) *CompiledFlowNode {
	t.Helper()

	entry, err := CompileCommand(FlowData{
		Nodes: []FlowNode{
			{ID: "entry", Type: FlowNodeTypeEntryCommand, Data: FlowNodeData{Name: "wait", Description: "Wait"}},
			{ID: "sleep", Type: FlowNodeTypeControlSleep, Data: FlowNodeData{SleepDurationSeconds: seconds}},
			{ID: "log", Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogMessage: "after"}},
		},
		Edges: []FlowEdge{
			{Source: "entry", Target: "sleep"},
			{Source: "sleep", Target: "log"},
		},
	})
	require.NoError(t, err)
	return entry
}

func TestSleepLongerThanThresholdSuspends(t *testing.T) {
	interaction := &discord.InteractionEvent{ID: 1, Token: "secret"}
	c, p := newSleepTestContext(t, &TestContextData{interaction: interaction})

	before := time.Now().UTC()
	require.NoError(t, sleepFlow(t, "3600").Execute(c))

	assert.Empty(t, p.log.messages, "blocks after the sleep ran before it was over")

	require.Len(t, p.resumePoint.created, 1)
	rp := p.resumePoint.created[0]
	assert.Equal(t, ResumePointTypeTimer, rp.Type)
	assert.Equal(t, "sleep", rp.NodeID)
	assert.Equal(t, "secret", rp.InteractionToken)
	assert.WithinDuration(t, before.Add(time.Hour), rp.ResumeAt, 5*time.Second)

	// The execution ends, so the interaction is deferred right away.
	assert.Equal(t, api.DeferredMessageInteractionWithSource, p.discord.response.Type)
}

func TestSleepDoesNotDeferRespondedInteraction(t *testing.T) {
	c, p := newSleepTestContext(t, &TestContextData{interaction: &discord.InteractionEvent{ID: 1}})
	p.discord.responded = true

	require.NoError(t, sleepFlow(t, "60").Execute(c))
	assert.Zero(t, p.discord.response.Type, "deferred an interaction that already had a response")
}

func TestResumeAfterSleepRunsFollowingBlocks(t *testing.T) {
	sleep := sleepFlow(t, "3600").FindChildWithID("sleep", true)
	require.NotNil(t, sleep)

	c, p := newSleepTestContext(t, &TestContextData{})
	require.NoError(t, sleep.ResumeAfterSleep(c))

	assert.Equal(t, []string{"after"}, p.log.messages)
	assert.Empty(t, p.resumePoint.created, "resuming slept again")
}

func TestSleepInsideLoopStaysInMemory(t *testing.T) {
	entry, err := CompileCommand(FlowData{
		Nodes: []FlowNode{
			{ID: "entry", Type: FlowNodeTypeEntryCommand, Data: FlowNodeData{Name: "wait", Description: "Wait"}},
			{ID: "loop", Type: FlowNodeTypeControlLoop, Data: FlowNodeData{LoopCount: "2"}},
			{ID: "each", Type: FlowNodeTypeControlLoopEach},
			{ID: "end", Type: FlowNodeTypeControlLoopEnd},
			{ID: "sleep", Type: FlowNodeTypeControlSleep, Data: FlowNodeData{SleepDurationSeconds: "60"}},
		},
		Edges: []FlowEdge{
			{Source: "entry", Target: "loop"},
			{Source: "loop", Target: "each"},
			{Source: "loop", Target: "end"},
			{Source: "each", Target: "sleep"},
		},
	})
	require.NoError(t, err)

	c, p := newSleepTestContext(t, &TestContextData{})
	err = entry.Execute(c)

	// 60 seconds don't fit into the execution, and the loop can't continue
	// after a durable sleep.
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sleep would exceed deadline")
	assert.Empty(t, p.resumePoint.created)
}

func TestSleepHugeDurationFailsInsteadOfOverflowing(t *testing.T) {
	c, p := newSleepTestContext(t, &TestContextData{})

	err := sleepFlow(t, "10000000000").Execute(c)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "can't be longer than")
	assert.Empty(t, p.log.messages, "overflowed duration continued the flow")
}

func TestSleepCountsDurableSleeps(t *testing.T) {
	c, p := newSleepTestContext(t, &TestContextData{})
	c.DurableSleeps = 3

	require.NoError(t, sleepFlow(t, "60").Execute(c))
	require.Len(t, p.resumePoint.created, 1)
	assert.Equal(t, 4, p.resumePoint.created[0].State.DurableSleeps)

	// A cycle through the sleep block stops at the limit.
	c, p = newSleepTestContext(t, &TestContextData{})
	c.DurableSleeps = maxDurableSleeps

	require.Error(t, sleepFlow(t, "60").Execute(c))
	assert.Empty(t, p.resumePoint.created)
}

func TestResumeAfterSleepRunsEnclosingErrorHandler(t *testing.T) {
	entry, err := CompileCommand(FlowData{
		Nodes: []FlowNode{
			{ID: "entry", Type: FlowNodeTypeEntryCommand, Data: FlowNodeData{Name: "wait", Description: "Wait"}},
			{ID: "handler", Type: FlowNodeTypeControlErrorHandler},
			{ID: "sleep", Type: FlowNodeTypeControlSleep, Data: FlowNodeData{SleepDurationSeconds: "60"}},
			// Fails because the expression is invalid.
			{ID: "fail", Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogMessage: "{{ ) }}"}},
			{ID: "caught", Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogMessage: "caught"}},
		},
		Edges: []FlowEdge{
			{Source: "entry", Target: "handler"},
			{Source: "handler", Target: "sleep"},
			{Source: "sleep", Target: "fail"},
			{Source: "handler", Target: "caught", SourceHandle: null.StringFrom("error")},
		},
	})
	require.NoError(t, err)

	c, p := newSleepTestContext(t, &TestContextData{})
	require.NoError(t, entry.FindChildWithID("sleep", true).ResumeAfterSleep(c))
	assert.Equal(t, []string{"caught"}, p.log.messages)
}

func TestResumeAfterSleepPassesErrorToOuterErrorHandler(t *testing.T) {
	entry, err := CompileCommand(FlowData{
		Nodes: []FlowNode{
			{ID: "entry", Type: FlowNodeTypeEntryCommand, Data: FlowNodeData{Name: "wait", Description: "Wait"}},
			{ID: "outer", Type: FlowNodeTypeControlErrorHandler},
			{ID: "inner", Type: FlowNodeTypeControlErrorHandler},
			{ID: "sleep", Type: FlowNodeTypeControlSleep, Data: FlowNodeData{SleepDurationSeconds: "60"}},
			{ID: "fail", Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogMessage: "{{ ) }}"}},
			// The inner handler's error branch fails too.
			{ID: "inner_fail", Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogMessage: "{{ ) }}"}},
			{ID: "caught", Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogMessage: "caught by outer"}},
		},
		Edges: []FlowEdge{
			{Source: "entry", Target: "outer"},
			{Source: "outer", Target: "inner"},
			{Source: "inner", Target: "sleep"},
			{Source: "sleep", Target: "fail"},
			{Source: "inner", Target: "inner_fail", SourceHandle: null.StringFrom("error")},
			{Source: "outer", Target: "caught", SourceHandle: null.StringFrom("error")},
		},
	})
	require.NoError(t, err)

	c, p := newSleepTestContext(t, &TestContextData{})
	require.NoError(t, entry.FindChildWithID("sleep", true).ResumeAfterSleep(c))
	assert.Equal(t, []string{"caught by outer"}, p.log.messages)
}

// buttonBranchSleepFlow sends a message with a button from inside wrapper and
// sleeps in the button's branch, which runs in a later execution.
func buttonBranchSleepFlow(t *testing.T, wrapper []FlowNode, wrapperEdges ...FlowEdge) *CompiledFlowNode {
	t.Helper()

	entry, err := CompileCommand(FlowData{
		Nodes: append([]FlowNode{
			{ID: "entry", Type: FlowNodeTypeEntryCommand, Data: FlowNodeData{Name: "wait", Description: "Wait"}},
			{ID: "message", Type: FlowNodeTypeActionMessageCreate},
			{ID: "sleep", Type: FlowNodeTypeControlSleep, Data: FlowNodeData{SleepDurationSeconds: "60"}},
			{ID: "fail", Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogMessage: "{{ ) }}"}},
		}, wrapper...),
		Edges: append([]FlowEdge{
			{Source: "entry", Target: wrapper[0].ID},
			{Source: "message", Target: "sleep", SourceHandle: null.StringFrom("component_1")},
			{Source: "sleep", Target: "fail"},
		}, wrapperEdges...),
	})
	require.NoError(t, err)
	return entry.FindChildWithID("sleep", true)
}

func TestSleepInButtonBranchIsNotInLoop(t *testing.T) {
	sleep := buttonBranchSleepFlow(t,
		[]FlowNode{
			{ID: "loop", Type: FlowNodeTypeControlLoop, Data: FlowNodeData{LoopCount: "2"}},
			{ID: "each", Type: FlowNodeTypeControlLoopEach},
			{ID: "end", Type: FlowNodeTypeControlLoopEnd},
		},
		FlowEdge{Source: "loop", Target: "each"},
		FlowEdge{Source: "loop", Target: "end"},
		FlowEdge{Source: "each", Target: "message"},
	)
	require.NotNil(t, sleep)
	assert.False(t, sleep.inLoop(), "the click that runs the branch doesn't loop")
}

func TestResumeInButtonBranchSkipsOuterErrorHandler(t *testing.T) {
	sleep := buttonBranchSleepFlow(t,
		[]FlowNode{{ID: "handler", Type: FlowNodeTypeControlErrorHandler}},
		FlowEdge{Source: "handler", Target: "message"},
	)
	require.NotNil(t, sleep)
	assert.Empty(t, sleep.enclosingErrorHandlers(), "the handler doesn't wrap the button branch")
}

func TestSleepNegativeDurationDoesNotSuspend(t *testing.T) {
	c, p := newSleepTestContext(t, &TestContextData{})

	require.NoError(t, sleepFlow(t, "-10000000000").Execute(c))
	assert.Empty(t, p.resumePoint.created)
	assert.Equal(t, []string{"after"}, p.log.messages)
}

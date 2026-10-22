package flow

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

// syncLogProvider records log messages from several branches at once.
type syncLogProvider struct {
	mu       sync.Mutex
	messages []string
}

func (p *syncLogProvider) CreateLogEntry(ctx context.Context, level provider.LogLevel, message string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.messages = append(p.messages, message)
}

func newParallelTestContext(t *testing.T, limits FlowContextLimits) (*FlowContext, *syncLogProvider, *recordingResumePointProvider) {
	t.Helper()

	log := &syncLogProvider{}
	resumePoint := &recordingResumePointProvider{}
	c := NewContext(
		context.Background(),
		5*time.Second,
		&TestContextData{},
		FlowProviders{Discord: &TestDiscordProvider{}, Log: log, ResumePoint: resumePoint},
		limits,
		eval.NewContext(eval.Env{}),
		nil,
	)
	t.Cleanup(c.Cancel)
	return c, log, resumePoint
}

var parallelTestLimits = FlowContextLimits{MaxStackDepth: 20, MaxOperations: 1000, MaxCredits: 1000}

// parallelFlow is a command with a Run in Parallel block, whose branches are
// the given nodes, each connected to the parallel block, and a log block
// after it that logs after.
func parallelFlow(t *testing.T, after string, branches []FlowNode, edges []FlowEdge) *CompiledFlowNode {
	t.Helper()

	nodes := []FlowNode{
		{ID: "entry", Type: FlowNodeTypeEntryCommand, Data: FlowNodeData{Name: "parallel", Description: "Parallel"}},
		{ID: "parallel", Type: FlowNodeTypeControlParallel},
		{ID: "after", Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogMessage: after}},
	}
	allEdges := []FlowEdge{
		{Source: "entry", Target: "parallel"},
		{Source: "parallel", Target: "after", SourceHandle: null.StringFrom("after")},
	}
	entry, err := CompileCommand(FlowData{
		Nodes: append(nodes, branches...),
		Edges: append(allEdges, edges...),
	})
	require.NoError(t, err)
	return entry
}

func TestParallelBranchesRunAtTheSameTime(t *testing.T) {
	entry := parallelFlow(t, "after",
		[]FlowNode{
			{ID: "sleep_a", Type: FlowNodeTypeControlSleep, Data: FlowNodeData{SleepDurationSeconds: "1"}},
			{ID: "log_a", Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogMessage: "a"}},
			{ID: "sleep_b", Type: FlowNodeTypeControlSleep, Data: FlowNodeData{SleepDurationSeconds: "1"}},
			{ID: "log_b", Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogMessage: "b"}},
		},
		[]FlowEdge{
			{Source: "parallel", Target: "sleep_a"},
			{Source: "sleep_a", Target: "log_a"},
			{Source: "parallel", Target: "sleep_b"},
			{Source: "sleep_b", Target: "log_b"},
		},
	)

	c, log, _ := newParallelTestContext(t, parallelTestLimits)
	start := time.Now()
	require.NoError(t, entry.Execute(c))

	assert.Less(t, time.Since(start), 1900*time.Millisecond, "branches ran one after the other")
	require.Len(t, log.messages, 3)
	assert.ElementsMatch(t, []string{"a", "b"}, log.messages[:2])
	assert.Equal(t, "after", log.messages[2], "the after output ran before all branches finished")
}

func TestParallelResultsAreAvailableAfter(t *testing.T) {
	entry := parallelFlow(t, "{{result('a')}} {{result('b')}}",
		[]FlowNode{
			{ID: "a", Type: FlowNodeTypeActionExpressionEvaluate, Data: FlowNodeData{Expression: "1 + 1"}},
			{ID: "b", Type: FlowNodeTypeActionExpressionEvaluate, Data: FlowNodeData{Expression: "2 + 2"}},
		},
		[]FlowEdge{
			{Source: "parallel", Target: "a"},
			{Source: "parallel", Target: "b"},
		},
	)

	c, log, _ := newParallelTestContext(t, parallelTestLimits)
	require.NoError(t, entry.Execute(c))
	assert.Equal(t, []string{"2 4"}, log.messages)
}

func TestParallelFailedBranchLetsOthersFinish(t *testing.T) {
	entry := parallelFlow(t, "after",
		[]FlowNode{
			{ID: "broken", Type: FlowNodeTypeActionExpressionEvaluate, Data: FlowNodeData{Expression: "1 +"}},
			{ID: "sleep", Type: FlowNodeTypeControlSleep, Data: FlowNodeData{SleepDurationSeconds: "1"}},
			{ID: "log", Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogMessage: "finished"}},
		},
		[]FlowEdge{
			{Source: "parallel", Target: "broken"},
			{Source: "parallel", Target: "sleep"},
			{Source: "sleep", Target: "log"},
		},
	)

	c, log, _ := newParallelTestContext(t, parallelTestLimits)
	err := entry.Execute(c)
	require.Error(t, err)
	assert.Equal(t, []string{"finished"}, log.messages, "the after output ran, or the other branch didn't finish")
}

func TestParallelSleepStaysInMemory(t *testing.T) {
	entry := parallelFlow(t, "after",
		[]FlowNode{
			{ID: "sleep", Type: FlowNodeTypeControlSleep, Data: FlowNodeData{SleepDurationSeconds: "60"}},
		},
		[]FlowEdge{
			{Source: "parallel", Target: "sleep"},
		},
	)

	c, _, resumePoint := newParallelTestContext(t, parallelTestLimits)
	err := entry.Execute(c)

	// The after output has to wait for the branch, which a resumed flow
	// couldn't do, so the 60 seconds have to fit into the execution.
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sleep would exceed deadline")
	assert.Empty(t, resumePoint.created)
}

func TestParallelSleepAfterBranchesSuspends(t *testing.T) {
	entry, err := CompileCommand(FlowData{
		Nodes: []FlowNode{
			{ID: "entry", Type: FlowNodeTypeEntryCommand, Data: FlowNodeData{Name: "parallel", Description: "Parallel"}},
			{ID: "parallel", Type: FlowNodeTypeControlParallel},
			{ID: "sleep", Type: FlowNodeTypeControlSleep, Data: FlowNodeData{SleepDurationSeconds: "60"}},
		},
		Edges: []FlowEdge{
			{Source: "entry", Target: "parallel"},
			{Source: "parallel", Target: "sleep", SourceHandle: null.StringFrom("after")},
		},
	})
	require.NoError(t, err)

	c, _, resumePoint := newParallelTestContext(t, parallelTestLimits)
	require.NoError(t, entry.Execute(c))
	assert.Len(t, resumePoint.created, 1)
}

func TestParallelBranchesShareLimits(t *testing.T) {
	var branches []FlowNode
	var edges []FlowEdge
	for _, id := range []string{"a", "b", "c", "d", "e"} {
		branches = append(branches, FlowNode{ID: id, Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogMessage: id}})
		edges = append(edges, FlowEdge{Source: "parallel", Target: id})
	}
	entry := parallelFlow(t, "after", branches, edges)

	// Each log block costs 1 credit, so only 3 of the 5 branches fit.
	c, _, _ := newParallelTestContext(t, FlowContextLimits{MaxStackDepth: 20, MaxOperations: 1000, MaxCredits: 3})
	err := entry.Execute(c)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "credits limit exceeded")
	assert.Equal(t, 5, c.CreditsUsed())
}

func TestParallelRejectsBlockInTwoBranches(t *testing.T) {
	entry := parallelFlow(t, "after",
		[]FlowNode{
			{ID: "a", Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogMessage: "a"}},
			{ID: "b", Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogMessage: "b"}},
			{ID: "shared", Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogMessage: "shared"}},
		},
		[]FlowEdge{
			{Source: "parallel", Target: "a"},
			{Source: "parallel", Target: "b"},
			{Source: "a", Target: "shared"},
			{Source: "b", Target: "shared"},
		},
	)

	c, log, _ := newParallelTestContext(t, parallelTestLimits)
	err := entry.Execute(c)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `block "shared" is connected to more than one branch`)
	assert.Empty(t, log.messages)
}

func TestParallelStateCanBeCopiedWhileBranchesRun(t *testing.T) {
	// Each branch runs a condition in a loop, while the state is copied like
	// when another branch sends a message with buttons. Fails with -race if
	// the condition or loop changes its state without the lock.
	var branches []FlowNode
	var edges []FlowEdge
	for _, id := range []string{"a", "b"} {
		branches = append(branches,
			FlowNode{ID: id + "_loop", Type: FlowNodeTypeControlLoop, Data: FlowNodeData{LoopCount: "50"}},
			FlowNode{ID: id + "_each", Type: FlowNodeTypeControlLoopEach},
			FlowNode{ID: id + "_end", Type: FlowNodeTypeControlLoopEnd},
			FlowNode{ID: id + "_condition", Type: FlowNodeTypeControlConditionCompare, Data: FlowNodeData{ConditionBaseValue: "1"}},
			FlowNode{ID: id + "_item", Type: FlowNodeTypeControlConditionItemCompare, Data: FlowNodeData{ConditionItemMode: ComparsionModeEqual, ConditionItemValue: "1"}},
		)
		edges = append(edges,
			FlowEdge{Source: "parallel", Target: id + "_loop"},
			FlowEdge{Source: id + "_loop", Target: id + "_each"},
			FlowEdge{Source: id + "_loop", Target: id + "_end"},
			FlowEdge{Source: id + "_each", Target: id + "_condition"},
			FlowEdge{Source: id + "_condition", Target: id + "_item"},
		)
	}
	entry := parallelFlow(t, "after", branches, edges)

	c, log, _ := newParallelTestContext(t, parallelTestLimits)
	done := make(chan struct{})
	copied := make(chan struct{})
	go func() {
		defer close(copied)
		for {
			select {
			case <-done:
				return
			default:
				c.FlowContextState.Copy()
			}
		}
	}()

	err := entry.Execute(c)
	close(done)
	<-copied
	require.NoError(t, err)
	assert.Equal(t, []string{"after"}, log.messages)
}

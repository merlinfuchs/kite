package flow

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"github.com/kitecloud/kite/kite-service/pkg/webhook"
)

// maxTestSteps bounds the blocks a test run records, a loop can run the same
// blocks many times.
const maxTestSteps = 300

// TestRun is set on the context of a test run, which runs a flow without
// changing anything. The providers of a test run record what blocks would do
// instead of doing it, and the run records the blocks that ran. Blocks that
// call services outside of Discord don't run, unless they are mocked.
type TestRun struct {
	mu sync.Mutex

	mocks     map[string]TestMock
	usedMocks map[string]bool
	steps     []TestStep
	truncated bool
	// The steps of the blocks that are running, innermost last. -1 for blocks
	// that weren't recorded.
	running []int
}

// TestMock replaces what a block does in a test run.
type TestMock struct {
	// Result is stored as the block's result, in the format of its result type,
	// like a Discord member object for Get Member.
	Result json.RawMessage
	// Error makes the block fail with it.
	Error string
}

// Err is the error the mock makes its block fail with, nil if none.
func (m TestMock) Err() error {
	if m.Error == "" {
		return nil
	}
	return &FlowError{Code: FlowNodeErrorUnknown, Message: m.Error}
}

// TestStep is a block that ran in a test run.
type TestStep struct {
	NodeID   string
	NodeType FlowNodeType
	// Null if the block has no result.
	Result    thing.Thing
	Error     string
	ErrorCode FlowNodeErrorCode
	// How the block ran differently than it would outside of a test, like
	// being mocked.
	Notes []string
	// What the block would have changed.
	Effects []TestEffect
}

// TestEffect is a change a block would have made, like sending a message.
type TestEffect struct {
	Action string
	// What would have been sent.
	Data any
	// Why Discord would likely reject it, like a missing permission.
	Warnings []string
}

func NewTestRun(mocks map[string]TestMock) *TestRun {
	return &TestRun{
		mocks:     mocks,
		usedMocks: make(map[string]bool),
	}
}

// Steps returns the blocks that ran, in the order they started.
func (t *TestRun) Steps() []TestStep {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.steps
}

// Truncated reports whether more blocks ran than were recorded.
func (t *TestRun) Truncated() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.truncated
}

// UsedMocks returns the IDs of the mocked blocks that ran.
func (t *TestRun) UsedMocks() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return slices.Collect(maps.Keys(t.usedMocks))
}

// AddEffect records a change of the block that is running.
func (t *TestRun) AddEffect(effect TestEffect) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if step := t.current(); step != nil {
		step.Effects = append(step.Effects, effect)
	}
}

// AddNote records how the block that is running ran differently.
func (t *TestRun) AddNote(note string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if step := t.current(); step != nil {
		step.Notes = append(step.Notes, note)
	}
}

// CurrentMock returns the mock of the block that is running, and notes that
// the block was mocked.
func (t *TestRun) CurrentMock() (TestMock, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	step := t.current()
	if step == nil {
		return TestMock{}, false
	}
	mock, ok := t.mocks[step.NodeID]
	if ok {
		t.usedMocks[step.NodeID] = true
		step.Notes = append(step.Notes, "Mocked")
	}
	return mock, ok
}

// current returns the step of the innermost running block, or the last step
// for changes made after a block, like the error response of a failed flow.
func (t *TestRun) current() *TestStep {
	for i := len(t.running) - 1; i >= 0; i-- {
		if t.running[i] >= 0 {
			return &t.steps[t.running[i]]
		}
	}
	if len(t.steps) > 0 {
		return &t.steps[len(t.steps)-1]
	}
	return nil
}

func (t *TestRun) enter(n *CompiledFlowNode) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(t.steps) >= maxTestSteps {
		t.truncated = true
		t.running = append(t.running, -1)
		return
	}

	t.steps = append(t.steps, TestStep{
		NodeID:   n.ID,
		NodeType: n.Type,
		Result:   thing.Null,
	})
	t.running = append(t.running, len(t.steps)-1)
}

func (t *TestRun) exit(n *CompiledFlowNode, ctx *FlowContext, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	i := t.running[len(t.running)-1]
	t.running = t.running[:len(t.running)-1]
	if i < 0 {
		return
	}

	step := &t.steps[i]
	step.Result = ctx.GetNodeResult(n.ID)
	if err != nil && step.Error == "" {
		// Errors pass through the blocks before the one they happened in.
		var trace *FlowErrorTrace
		if errors.As(err, &trace) && trace.NodeID == n.ID {
			step.Error = trace.Next.Error()
			step.ErrorCode = DescribeError(err).Code
		}
	}

	// Every case of a condition runs to check if it matches, only the ones
	// that ran blocks are worth showing.
	if n.IsConditionItem() && i == len(t.steps)-1 && step.Error == "" && len(step.Effects) == 0 && len(step.Notes) == 0 {
		t.steps = t.steps[:i]
	}
}

func (t *TestRun) run(n *CompiledFlowNode, ctx *FlowContext) error {
	t.enter(n)
	err := t.runBlock(n, ctx)
	t.exit(n, ctx, err)
	return err
}

func (t *TestRun) runBlock(n *CompiledFlowNode, ctx *FlowContext) error {
	if n.Type == FlowNodeTypeControlWebhookVerify {
		return t.runWebhookVerify(n, ctx)
	}
	if !n.IsAction() {
		return n.runBlock(ctx)
	}

	mock, mocked := t.CurrentMock()
	if !mocked && !n.callsExternalService() {
		return n.runBlock(ctx)
	}

	if err := n.checkIntegrations(ctx); err != nil {
		return traceError(n, err)
	}

	if !mocked {
		t.AddNote("Not run, test runs don't send requests outside of Discord. Mock the block to set its result.")
		return n.ExecuteChildren(ctx)
	}

	if err := mock.Err(); err != nil {
		return traceError(n, err)
	}
	if len(mock.Result) > 0 {
		result, err := TestMockResult(n.Type, mock.Result)
		if err != nil {
			return traceError(n, fmt.Errorf("invalid mock result: %w", err))
		}
		ctx.StoreNodeResult(n, result)
	}
	return n.ExecuteChildren(ctx)
}

// runWebhookVerify takes the valid branch, unless the block is mocked with
// "invalid". The secrets of test runs aren't the real ones, so signatures
// can't be checked.
func (t *TestRun) runWebhookVerify(n *CompiledFlowNode, ctx *FlowContext) error {
	if _, ok := ctx.Data.Event().(*webhook.Event); !ok {
		return traceError(n, errors.New("only webhook event listeners can verify a webhook request"))
	}

	handle := "valid"
	if mock, ok := t.CurrentMock(); ok {
		if err := mock.Err(); err != nil {
			return traceError(n, err)
		}
		if err := json.Unmarshal(mock.Result, &handle); err != nil || (handle != "valid" && handle != "invalid") {
			return traceError(n, errors.New(`invalid mock result: must be "valid" or "invalid"`))
		}
	} else {
		t.AddNote(`Signature not checked in test runs, took the valid branch. Mock the block with "invalid" to take the other one.`)
	}
	return n.ExecuteChildrenByHandle(ctx, handle)
}

// callsExternalService reports whether n sends requests to a service other
// than Discord, which test runs don't do.
func (n *CompiledFlowNode) callsExternalService() bool {
	switch n.Type {
	case FlowNodeTypeActionHTTPRequest,
		FlowNodeTypeActionRobloxUserGet,
		FlowNodeTypeActionAIChatCompletion,
		FlowNodeTypeActionAISearchWeb,
		FlowNodeTypeActionAIScore:
		return true
	}

	if block, ok := blockDefinitions[n.Type]; ok && block.Run.Kind == "request" {
		return integrations[block.Run.Integration].Auth.Type != "discord_bot"
	}
	return false
}

// TestMockResult decodes the mocked result of a block of type t, in the format
// of the result type of the block.
func TestMockResult(t FlowNodeType, raw json.RawMessage) (thing.Thing, error) {
	result := blockDefinitions[t].Result
	if result == nil || result.Thing == "" {
		return discordAPIResult(raw)
	}

	decode := func(item json.RawMessage) (thing.Thing, error) {
		if result.Thing == string(thing.TypeHTTPResponse) {
			return decodeMockHTTPResponse(item)
		}
		// Like a stored result, which has its type next to its value.
		var v thing.Thing
		err := json.Unmarshal(fmt.Appendf(nil, `{"t":%q,"v":%s}`, result.Thing, item), &v)
		return v, err
	}
	if !result.List {
		return decode(raw)
	}

	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		return thing.Null, fmt.Errorf("must be a list: %w", err)
	}
	res := make([]thing.Thing, len(items))
	for i, item := range items {
		v, err := decode(item)
		if err != nil {
			return thing.Null, err
		}
		res[i] = v
	}
	return thing.NewArray(res), nil
}

// decodeMockHTTPResponse decodes {"status_code": 200, "body": ...}, where the
// body is text, or JSON that the response would have contained.
func decodeMockHTTPResponse(raw json.RawMessage) (thing.Thing, error) {
	var v struct {
		StatusCode int               `json:"status_code"`
		Body       json.RawMessage   `json:"body"`
		Headers    map[string]string `json:"headers"`
	}
	if err := json.Unmarshal(raw, &v); err != nil {
		return thing.Null, err
	}
	if v.StatusCode == 0 {
		v.StatusCode = http.StatusOK
	}

	body := []byte(v.Body)
	var text string
	if json.Unmarshal(v.Body, &text) == nil {
		body = []byte(text)
	}

	return thing.NewHTTPResponse(thing.HTTPResponseValue{
		Status:     fmt.Sprintf("%d %s", v.StatusCode, http.StatusText(v.StatusCode)),
		StatusCode: v.StatusCode,
		Body:       bytes.TrimSpace(body),
		Headers:    v.Headers,
	}), nil
}

// TestSnowflake returns a new ID for things test runs pretend to create, like
// sent messages.
var TestSnowflake = func() func() discord.Snowflake {
	var mu sync.Mutex
	var last discord.Snowflake
	return func() discord.Snowflake {
		mu.Lock()
		defer mu.Unlock()
		id := discord.NewSnowflake(time.Now())
		if id <= last {
			id = last + 1
		}
		last = id
		return id
	}
}()

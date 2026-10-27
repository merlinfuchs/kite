package flowtest

import (
	"context"
	"errors"
	"fmt"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/core/engine"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

type Runner interface {
	Run(ctx context.Context, appID string, test engine.FlowTest) (*engine.FlowTestResult, error)
}

type FlowTestHandler struct {
	runner    Runner
	userStore store.UserStore
}

func NewFlowTestHandler(runner Runner, userStore store.UserStore) *FlowTestHandler {
	return &FlowTestHandler{
		runner:    runner,
		userStore: userStore,
	}
}

func (h *FlowTestHandler) HandleCommandTest(c *handler.Context, req wire.FlowTestRequest) (*wire.FlowTestResponse, error) {
	source := c.Command.FlowSource
	if req.FlowSource != nil {
		source = *req.FlowSource
	}
	return h.run(c, req, engine.FlowTest{Kind: engine.FlowTestKindCommand, Flow: source})
}

func (h *FlowTestHandler) HandleEventListenerTest(c *handler.Context, req wire.FlowTestRequest) (*wire.FlowTestResponse, error) {
	source := c.EventListener.FlowSource
	if req.FlowSource != nil {
		source = *req.FlowSource
	}
	return h.run(c, req, engine.FlowTest{Kind: engine.FlowTestKindEventListener, Flow: source})
}

func (h *FlowTestHandler) HandleMessageTest(c *handler.Context, req wire.FlowTestRequest) (*wire.FlowTestResponse, error) {
	if req.FlowSource != nil {
		return nil, handler.ErrBadRequest("invalid_test", "Messages are tested as they are saved, without flow_source")
	}
	if req.FlowSourceID == "" {
		return nil, handler.ErrBadRequest("invalid_test", "flow_source_id of the button or select menu to test is required")
	}
	return h.run(c, req, engine.FlowTest{
		Kind:         engine.FlowTestKindMessage,
		MessageID:    c.Message.ID,
		FlowSourceID: req.FlowSourceID,
	})
}

func (h *FlowTestHandler) run(c *handler.Context, req wire.FlowTestRequest, test engine.FlowTest) (*wire.FlowTestResponse, error) {
	// Tests change nothing, but read the app's servers through the bot.
	if c.APIToken != nil {
		if err := handler.RequireAPITokenAccess(c.APIToken, model.APITokenResourceKite, model.APITokenAccessRead); err != nil {
			return nil, err
		}
		if err := handler.RequireAPITokenAccess(c.APIToken, model.APITokenResourceDiscord, model.APITokenAccessRead); err != nil {
			return nil, err
		}
	}

	test.UserID = req.UserID
	if test.UserID == "" {
		user, err := h.userStore.User(c.Context(), c.Session.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}
		test.UserID = user.DiscordID
	}

	test.GuildID = req.GuildID
	test.ChannelID = req.ChannelID
	test.DM = req.DM
	test.Options = req.Options
	test.Values = req.Values
	test.Event = req.Event
	if req.Webhook != nil {
		test.Webhook = &engine.FlowTestWebhook{
			Body:    req.Webhook.Body,
			Headers: req.Webhook.Headers,
			Query:   req.Webhook.Query,
		}
	}

	test.Mocks = make(map[string]flow.TestMock, len(req.Mocks))
	for id, mock := range req.Mocks {
		if len(mock.Result) == 0 && mock.Error == "" {
			return nil, handler.ErrBadRequest("invalid_test", fmt.Sprintf("Mock of block %s needs a result or an error", id))
		}
		test.Mocks[id] = flow.TestMock{Result: mock.Result, Error: mock.Error}
	}
	for _, action := range req.Then {
		test.Then = append(test.Then, engine.FlowTestAction{
			Click:  action.Click,
			Select: action.Select,
			Values: action.Values,
			Submit: action.Submit,
			UserID: action.UserID,
		})
	}

	res, err := h.runner.Run(c.Context(), c.App.ID, test)
	if err != nil {
		var testErr *engine.FlowTestError
		switch {
		case errors.As(err, &testErr):
			return nil, handler.ErrBadRequest("invalid_test", testErr.Message)
		case errors.Is(err, engine.ErrTestAppOffline):
			return nil, handler.ErrBadRequest("app_not_running", "The app isn't running, enable it first")
		}
		return nil, fmt.Errorf("failed to run test: %w", err)
	}

	return flowTestResponse(res), nil
}

func flowTestResponse(res *engine.FlowTestResult) *wire.FlowTestResponse {
	resp := &wire.FlowTestResponse{
		Executions:  make([]wire.FlowTestExecution, len(res.Executions)),
		Credits:     res.Credits,
		UnusedMocks: res.UnusedMocks,
	}

	for i, exec := range res.Executions {
		steps := make([]wire.FlowTestStep, len(exec.Steps))
		for j, step := range exec.Steps {
			effects := make([]wire.FlowTestEffect, len(step.Effects))
			for k, effect := range step.Effects {
				effects[k] = wire.FlowTestEffect{
					Action:   effect.Action,
					Data:     effect.Data,
					Warnings: effect.Warnings,
				}
			}

			var result any
			if !step.Result.IsNil() {
				result = thingValue(step.Result)
			}
			steps[j] = wire.FlowTestStep{
				NodeID:    step.NodeID,
				NodeType:  string(step.NodeType),
				Result:    result,
				Error:     step.Error,
				ErrorCode: string(step.ErrorCode),
				Notes:     step.Notes,
				Effects:   effects,
			}
		}

		resp.Executions[i] = wire.FlowTestExecution{
			Trigger:   exec.Trigger,
			Steps:     steps,
			Truncated: exec.Truncated,
			Notes:     exec.Notes,
			Error:     exec.Error,
		}
	}
	return resp
}

// thingValue is the value of a result as plain JSON, without the types of the
// things in lists and objects.
func thingValue(t thing.Thing) any {
	switch t.Type {
	case thing.TypeArray:
		items := t.Array()
		res := make([]any, len(items))
		for i, item := range items {
			res[i] = thingValue(item)
		}
		return res
	case thing.TypeObject:
		fields := t.Object()
		res := make(map[string]any, len(fields))
		for k, v := range fields {
			res[k] = thingValue(v)
		}
		return res
	case thing.TypeHTTPResponse:
		resp := t.HTTPResponse()
		return map[string]any{
			"status_code": resp.StatusCode,
			"body":        string(resp.Body),
		}
	}
	return t.Value
}

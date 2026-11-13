package flow

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/kitecloud/kite/kite-service/pkg/provider"
)

// operationCall is a call of a handler operation from code.
type operationCall struct {
	// A copy of the flow's context that's canceled with the call.
	ctx *FlowContext
	arg json.RawMessage
	// charge counts the call against the flow's operations and charges its
	// credits. A handler calls it once the argument is valid, before it does
	// anything, so a wrong argument isn't billed.
	charge func() error
	// chargeCredits is charge for operations whose credits depend on the
	// argument, like the model of ai.ask. It charges credits instead of the
	// operation's, if credits and reserve fit the limit. reserve is the most
	// chargeMore can charge later.
	chargeCredits func(credits int, reserve int) error
	// chargeMore charges credits on top, for what's only known after paying
	// for something, like the length of an AI answer.
	chargeMore func(credits int) error
	// The IDs of the stored variables the block lets the code use, by the
	// names the code uses.
	variables map[string]string
}

// operationHandler runs a handler operation, see
// kite-web/src/lib/integrations/handlerOperations.ts, and returns its result,
// which the code gets as JSON.
type operationHandler func(call operationCall) (any, error)

// operationHandlers are the handlers of the handler operations, by their
// name, like http.request. TestEveryHandlerOperationHasAHandler checks they
// match operations.json.
var operationHandlers = map[string]operationHandler{
	"http.request":        httpRequestOperation,
	"variables.get":       variableGetOperation,
	"variables.set":       variableValueOperation(provider.VariableOperationOverwrite),
	"variables.append":    variableValueOperation(provider.VariableOperationAppend),
	"variables.prepend":   variableValueOperation(provider.VariableOperationPrepend),
	"variables.increment": variableIncrementOperation,
	"variables.delete":    variableDeleteOperation,
	"variables.list":      variableListOperation,
	"ai.ask":              aiAskOperation,
	"ai.check":            aiCheckOperation,
	"ai.choose":           aiChooseOperation,
	"ai.score":            aiScoreOperation,
	"ai.translate":        aiTranslateOperation,
}

// decodeOperationArg decodes the argument of a call into v, which has a field
// for each of its parameters.
func decodeOperationArg(arg json.RawMessage, v any) error {
	if len(bytes.TrimSpace(arg)) == 0 {
		return nil
	}
	dec := json.NewDecoder(bytes.NewReader(arg))
	dec.DisallowUnknownFields()
	err := dec.Decode(v)
	if err == nil {
		return nil
	}

	var typeErr *json.UnmarshalTypeError
	switch {
	case errors.As(err, &typeErr) && typeErr.Field != "":
		return fmt.Errorf("invalid value for parameter %s", typeErr.Field)
	case strings.HasPrefix(err.Error(), "json: unknown field "):
		return fmt.Errorf("it has no parameter %s", strings.Trim(strings.TrimPrefix(err.Error(), "json: unknown field "), `"`))
	}
	return errors.New("the argument must be an object of its parameters")
}

package provider

import (
	"testing"

	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"github.com/stretchr/testify/assert"
)

func TestVariableOperationApply(t *testing.T) {
	// A missing value counts as 0 for decrement, others store the value.
	assert.Equal(t, float64(-50), VariableOperationDecrement.Apply(thing.Null, false, thing.NewInt(50)).Float())
	assert.Equal(t, float64(-50), VariableOperationDecrement.Apply(thing.Null, false, thing.NewString("50")).Float())
	assert.Equal(t, int64(50), VariableOperationIncrement.Apply(thing.Null, false, thing.NewInt(50)).Int())
	assert.Equal(t, "hi", VariableOperationAppend.Apply(thing.Null, false, thing.NewString("hi")).String())
	// An object starts a list, so the next append adds to it.
	first := VariableOperationAppend.Apply(thing.Null, false, thing.NewObject(map[string]thing.Thing{"id": thing.NewInt(1)}))
	assert.Len(t, VariableOperationAppend.Apply(first, true, thing.NewObject(map[string]thing.Thing{"id": thing.NewInt(2)})).Array(), 2)

	assert.Equal(t, float64(15), VariableOperationIncrement.Apply(thing.NewInt(10), true, thing.NewInt(5)).Float())
	assert.Equal(t, "ab", VariableOperationAppend.Apply(thing.NewString("a"), true, thing.NewString("b")).String())
	assert.Equal(t, "ba", VariableOperationPrepend.Apply(thing.NewString("a"), true, thing.NewString("b")).String())
	assert.Equal(t, "b", VariableOperationOverwrite.Apply(thing.NewString("a"), true, thing.NewString("b")).String())
}

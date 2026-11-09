package postgres

import (
	"strings"
	"testing"

	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"github.com/stretchr/testify/assert"
)

func TestFirstVariableValue(t *testing.T) {
	assert.Equal(t, float64(-50), firstVariableValue(provider.VariableOperationDecrement, thing.NewInt(50)).Float())
	assert.Equal(t, float64(-50), firstVariableValue(provider.VariableOperationDecrement, thing.NewString("50")).Float())
	assert.Equal(t, int64(50), firstVariableValue(provider.VariableOperationIncrement, thing.NewInt(50)).Int())
	assert.Equal(t, "hi", firstVariableValue(provider.VariableOperationAppend, thing.NewString("hi")).String())
}

func TestEncodeVariableValueLimit(t *testing.T) {
	_, err := encodeVariableValue(thing.NewString(strings.Repeat("a", 60<<10)))
	assert.NoError(t, err)
	_, err = encodeVariableValue(thing.NewString(strings.Repeat("a", 64<<10)))
	assert.ErrorIs(t, err, provider.ErrVariableValueTooLarge)
}

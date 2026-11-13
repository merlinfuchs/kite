package postgres

import (
	"strings"
	"testing"

	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"github.com/stretchr/testify/assert"
)

func TestEncodeVariableValueLimit(t *testing.T) {
	_, err := encodeVariableValue(thing.NewString(strings.Repeat("a", 60<<10)))
	assert.NoError(t, err)
	_, err = encodeVariableValue(thing.NewString(strings.Repeat("a", 64<<10)))
	assert.ErrorIs(t, err, provider.ErrVariableValueTooLarge)
}

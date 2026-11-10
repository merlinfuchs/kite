package util

import (
	"testing"

	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/stretchr/testify/assert"
)

func TestNameStyleApplied(t *testing.T) {
	set := provider.NameStyle{FontID: 8, EffectID: 2, Colors: []int{1, 2}}
	assert.True(t, nameStyleApplied(set, &provider.NameStyle{FontID: 8, EffectID: 2, Colors: []int{1, 2}}))
	// Discord keeps the old style when it doesn't take the new one.
	assert.False(t, nameStyleApplied(set, &provider.NameStyle{FontID: 13, EffectID: 7, Colors: []int{1, 2}}))
	assert.False(t, nameStyleApplied(set, nil))

	var reset provider.NameStyle
	assert.True(t, nameStyleApplied(reset, nil))
	assert.False(t, nameStyleApplied(reset, &provider.NameStyle{FontID: 8, EffectID: 1, Colors: []int{1}}))
}

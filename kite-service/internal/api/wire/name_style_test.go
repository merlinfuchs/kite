package wire

import (
	"testing"

	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/stretchr/testify/assert"
)

func TestNameStyleValidate(t *testing.T) {
	assert.NoError(t, NameStyle{Font: "8bit", Effect: "prism", Colors: []int{0, 0xFFFFFF}}.Validate())
	assert.Error(t, NameStyle{Font: "comic_sans", Effect: "solid", Colors: []int{1}}.Validate())
	assert.Error(t, NameStyle{Font: "8bit", Effect: "glow", Colors: []int{1}}.Validate())
	assert.Error(t, NameStyle{Font: "8bit", Effect: "solid"}.Validate())
	assert.Error(t, NameStyle{Font: "8bit", Effect: "solid", Colors: []int{1, 2, 3, 4, 5, 6}}.Validate())
	assert.Error(t, NameStyle{Font: "8bit", Effect: "solid", Colors: []int{0x1000000}}.Validate())
}

// Styles with fonts Discord no longer offers have no name in Kite.
func TestNameStyleToWire(t *testing.T) {
	assert.Equal(t,
		&NameStyle{Font: "monkey_bars", Effect: "prism", Colors: []int{1, 2}},
		NameStyleToWire(&provider.NameStyle{FontID: 13, EffectID: 7, Colors: []int{1, 2}}),
	)
	assert.Nil(t, NameStyleToWire(&provider.NameStyle{FontID: 1, EffectID: 1, Colors: []int{1}}))
	assert.Nil(t, NameStyleToWire(nil))
}

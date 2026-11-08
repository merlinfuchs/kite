package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPITokenHasAccess(t *testing.T) {
	read := &APIToken{Scopes: []string{APITokenScopeKiteRead}}
	assert.True(t, read.HasAccess(APITokenResourceKite, APITokenAccessRead))
	assert.False(t, read.HasAccess(APITokenResourceKite, APITokenAccessWrite))

	write := &APIToken{Scopes: []string{APITokenScopeKiteWrite}}
	assert.True(t, write.HasAccess(APITokenResourceKite, APITokenAccessRead))
	assert.True(t, write.HasAccess(APITokenResourceKite, APITokenAccessWrite))
	assert.False(t, write.HasAccess("discord", APITokenAccessRead))

	// The scope of tokens before scopes were split.
	all := &APIToken{Scopes: []string{"all"}}
	assert.False(t, all.HasAccess(APITokenResourceKite, APITokenAccessRead))
}

func TestAPITokenCanUseApp(t *testing.T) {
	assert.True(t, (&APIToken{}).CanUseApp("a"))
	assert.True(t, (&APIToken{AppIDs: []string{"a"}}).CanUseApp("a"))
	assert.False(t, (&APIToken{AppIDs: []string{"a"}}).CanUseApp("b"))
	assert.False(t, (&APIToken{AppIDs: []string{}}).CanUseApp("a"))
}

func TestAPITokenAccessForMethod(t *testing.T) {
	assert.Equal(t, APITokenAccessRead, APITokenAccessForMethod("GET"))
	for _, method := range []string{"POST", "PUT", "PATCH", "DELETE"} {
		assert.Equal(t, APITokenAccessWrite, APITokenAccessForMethod(method))
	}
}

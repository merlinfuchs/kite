package webhook

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponseValidate(t *testing.T) {
	valid := []Response{
		{Status: 200, ContentType: "text/plain", Body: "ok"},
		{Status: 204, ContentType: "text/plain"},
		{Status: 401, ContentType: "application/json", Body: `{}`},
		{Status: 599, ContentType: "text/plain"},
		{Status: 200, ContentType: "text/plain", Body: strings.Repeat("a", MaxBodySize)},
	}
	for _, resp := range valid {
		assert.NoError(t, resp.Validate(), "%+v", resp)
	}

	invalid := []Response{
		{Status: 101, ContentType: "text/plain"},
		{Status: 302, ContentType: "text/plain"},
		{Status: 399, ContentType: "text/plain"},
		{Status: 600, ContentType: "text/plain"},
		{Status: 0, ContentType: "text/plain"},
		{Status: 204, ContentType: "text/plain", Body: "a"},
		{Status: 200, ContentType: "text/html"},
		{Status: 200, ContentType: "application/json", Body: `{"a":"b"c"}`},
		{Status: 200, ContentType: ""},
		{Status: 200, ContentType: "text/plain", Body: strings.Repeat("a", MaxBodySize+1)},
	}
	for _, resp := range invalid {
		assert.Error(t, resp.Validate(), "%+v", resp)
	}
}

func TestResponderRespondsOnce(t *testing.T) {
	r := NewResponder()
	require.NoError(t, r.Respond(&Response{Status: 200, ContentType: "text/plain"}))
	assert.ErrorIs(t, r.Respond(&Response{Status: 400, ContentType: "text/plain"}), ErrNotWaiting)

	resp := r.Wait(context.Background(), time.Second)
	require.NotNil(t, resp)
	assert.Equal(t, 200, resp.Status)
}

func TestResponderClose(t *testing.T) {
	r := NewResponder()
	r.Close()
	assert.Nil(t, r.Wait(context.Background(), time.Second))
	assert.ErrorIs(t, r.Respond(&Response{Status: 200, ContentType: "text/plain"}), ErrNotWaiting)
}

func TestResponderTimeout(t *testing.T) {
	r := NewResponder()
	assert.Nil(t, r.Wait(context.Background(), time.Millisecond))
	assert.ErrorIs(t, r.Respond(&Response{Status: 200, ContentType: "text/plain"}), ErrNotWaiting)
}

func TestResponderValidates(t *testing.T) {
	r := NewResponder()
	assert.Error(t, r.Respond(&Response{Status: 302, ContentType: "text/plain"}))

	// The invalid response doesn't answer the request.
	assert.NoError(t, r.Respond(&Response{Status: 200, ContentType: "text/plain"}))
}

func TestResponderNil(t *testing.T) {
	var r *Responder
	assert.ErrorIs(t, r.Respond(&Response{Status: 200, ContentType: "text/plain"}), ErrNotWaiting)
	r.Close()
}

package lookup

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleUserGet(t *testing.T) {
	discordCalls := 0
	discord := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		discordCalls++
		assert.Equal(t, "Bot token", r.Header.Get("Authorization"))

		switch r.URL.Path {
		case "/users/80351110224678912":
			w.Write([]byte(`{
				"id": "80351110224678912",
				"username": "nelly",
				"global_name": "Nelly",
				"discriminator": "0",
				"avatar": "8342729096ea3675442027381ff50dfe",
				"banner": null,
				"accent_color": 16711680,
				"public_flags": 64,
				"avatar_decoration_data": {"asset": "a_fed43ab12698df65902ba06727e20c0e", "sku_id": "1"},
				"primary_guild": {"identity_guild_id": "613425648685547541", "identity_enabled": true, "tag": "DDEV", "badge": "abc"}
			}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message": "Unknown User", "code": 10013}`))
		}
	}))
	defer discord.Close()

	h, err := NewLookupHandler("token")
	require.NoError(t, err)
	h.baseURL = discord.URL

	mux := http.NewServeMux()
	mux.Handle("GET /users/{userID}", handler.APIHandler(handler.Typed(h.HandleUserGet)))

	get := func(id string) (int, wire.APIResponse[*wire.LookupUserGetResponse]) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/users/"+id, nil))
		var resp wire.APIResponse[*wire.LookupUserGetResponse]
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		return rec.Code, resp
	}

	status, resp := get("80351110224678912")
	require.Equal(t, http.StatusOK, status)
	user := resp.Data
	assert.Equal(t, "nelly", user.Username)
	assert.Equal(t, "Nelly", user.GlobalName.String)
	assert.Equal(t, int64(16711680), user.AccentColor.Int64)
	assert.False(t, user.Banner.Valid)
	assert.Equal(t, "a_fed43ab12698df65902ba06727e20c0e", user.AvatarDecoration.String)
	require.NotNil(t, user.PrimaryGuild)
	assert.Equal(t, "DDEV", user.PrimaryGuild.Tag)

	status, resp = get("123456789012345678")
	assert.Equal(t, http.StatusNotFound, status)
	assert.Equal(t, "unknown_user", resp.Error.Code)

	status, resp = get("abc")
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, "invalid_user_id", resp.Error.Code)

	// Both answers are cached, so Discord isn't asked again.
	h.cache.Wait()
	get("80351110224678912")
	get("123456789012345678")
	assert.Equal(t, 2, discordCalls)
}

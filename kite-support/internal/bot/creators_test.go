package bot

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
)

func TestFetchCreators(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"success": true, "data": [
			{"id": "moderation", "owner": null},
			{"id": "a", "owner": {"discord_id": "111"}},
			{"id": "b", "owner": {"discord_id": "222"}},
			{"id": "c", "owner": {"discord_id": "111"}}
		]}`))
	}))
	defer srv.Close()

	creators, err := fetchCreators(context.Background(), srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	want := []discord.UserID{111, 222}
	if len(creators) != len(want) || creators[0] != want[0] || creators[1] != want[1] {
		t.Fatalf("got %v, want %v", creators, want)
	}
}

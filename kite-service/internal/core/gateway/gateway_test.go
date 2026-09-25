package gateway

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/util"
)

// Run with -race: restarts replace the session while the API, the manager and
// the engine's scheduler read it.
func TestReplaceSessionWhileReading(t *testing.T) {
	crypt, err := util.NewSymmetricCrypt(strings.Repeat("ab", 32))
	if err != nil {
		t.Fatal(err)
	}
	token, err := crypt.EncryptString("token")
	if err != nil {
		t.Fatal(err)
	}

	app := &model.App{ID: "app", DiscordToken: token}
	session, err := createSession(crypt, app)
	if err != nil {
		t.Fatal(err)
	}

	g := &Gateway{tokenCrypt: crypt, app: app, session: session}
	g.ctx, g.cancel = context.WithCancel(context.Background())

	var wg sync.WaitGroup
	for range 4 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			for range 50 {
				if _, _, err := g.replaceSession(); err != nil {
					t.Error(err)
				}
			}
		}()
		go func() {
			defer wg.Done()
			for range 50 {
				if g.Session() == nil {
					t.Error("no session")
				}
				_, _ = g.AppStatus(context.Background())
			}
		}()
	}
	wg.Wait()

	if g.Session() == session {
		t.Fatal("session wasn't replaced")
	}
}

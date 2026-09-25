package gateway

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/util"
)

// newTestGateway builds a gateway without connecting it.
func newTestGateway(t *testing.T) *Gateway {
	t.Helper()

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

	g := &Gateway{tokenCrypt: crypt, appID: app.ID, app: app, session: session}
	g.ctx, g.cancel = context.WithCancel(context.Background())
	return g
}

// Run with -race: restarts replace the session while the API, the manager and
// the engine's scheduler read it.
func TestReplaceSessionWhileReading(t *testing.T) {
	g := newTestGateway(t)
	first := g.Session()

	var wg sync.WaitGroup
	for range 4 {
		wg.Add(3)
		go func() {
			defer wg.Done()
			for range 50 {
				if _, _, err := g.replaceSession(nil); err != nil {
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
		go func() {
			defer wg.Done()
			for range 50 {
				g.rotatePresence(context.Background(), time.Now(), true)
			}
		}()
	}
	wg.Wait()

	if g.Session() == first {
		t.Fatal("session wasn't replaced")
	}
}

func TestReplacedSessionCancelsItsContext(t *testing.T) {
	g := newTestGateway(t)
	oldCtx := g.ctx

	_, ctx, err := g.replaceSession(nil)
	if err != nil {
		t.Fatal(err)
	}

	// startGateway of the old connection sees this and doesn't disable the app.
	if oldCtx.Err() == nil {
		t.Fatal("old connection's context wasn't cancelled")
	}
	if ctx.Err() != nil {
		t.Fatal("new connection's context is cancelled")
	}
}

func TestReplaceSessionSkipsWhenAlreadyReplaced(t *testing.T) {
	g := newTestGateway(t)
	stale := g.Session()

	if _, _, err := g.replaceSession(nil); err != nil {
		t.Fatal(err)
	}
	current := g.Session()

	// An intent refresh that started before the restart above.
	session, _, err := g.replaceSession(stale)
	if err != nil {
		t.Fatal(err)
	}
	if session != nil || g.Session() != current {
		t.Fatal("replaced a session that was already replaced")
	}
}

func TestClosedGatewayDoesNotReconnect(t *testing.T) {
	g := newTestGateway(t)
	if err := g.Close(); err != nil {
		t.Fatal(err)
	}

	session, _, err := g.replaceSession(nil)
	if err != nil {
		t.Fatal(err)
	}
	if session != nil {
		t.Fatal("closed gateway created a new session")
	}
}

func TestSendPresenceBeforeConnecting(t *testing.T) {
	g := newTestGateway(t)

	// The session hasn't opened its gateway yet, which used to panic.
	err := g.sendPresence(context.Background(), presenceForApp(g.currentApp()))
	if !errors.Is(err, errNotConnected) {
		t.Fatalf("err = %v, want errNotConnected", err)
	}
}

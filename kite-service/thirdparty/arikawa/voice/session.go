package voice

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/internal/moreatomic"
	"github.com/diamondburned/arikawa/v3/session"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/handler"
	"github.com/diamondburned/arikawa/v3/utils/ws"
	"github.com/diamondburned/arikawa/v3/utils/ws/ophandler"
	"github.com/diamondburned/arikawa/v3/voice/udp"
	"github.com/diamondburned/arikawa/v3/voice/voicegateway"
	"github.com/disgoorg/godave"
	davesession "github.com/thomas-vilte/dave-go/session"
)

// Protocol is the encryption protocol that this library uses.
const Protocol = "aead_xchacha20_poly1305_rtpsize"

// ErrAlreadyConnecting is returned when the session is already connecting.
var ErrAlreadyConnecting = errors.New("already connecting")

// ErrCannotSend is an error when audio is sent to a closed channel.
var ErrCannotSend = errors.New("cannot send audio to closed channel")

// WSTimeout is the duration to wait for a gateway operation including Session
// to complete before erroring out. This only applies to functions that don't
// take in a context already.
const WSTimeout = 25 * time.Second

// ReconnectError is emitted into Session.Handler everytime the voice gateway
// fails to be reconnected. It implements the error interface.
type ReconnectError struct {
	Err error
}

// Error implements error.
func (e ReconnectError) Error() string {
	return "voice reconnect error: " + e.Err.Error()
}

// Unwrap returns e.Err.
func (e ReconnectError) Unwrap() error { return e.Err }

// MainSession abstracts both session.Session and state.State.
type MainSession interface {
	// AddHandler describes the method in handler.Handler.
	AddHandler(handler any) (rm func())
	// Me returns the current user.
	Me() (*discord.User, error)
	// Channel queries for the channel with the given ID.
	Channel(discord.ChannelID) (*discord.Channel, error)
	// SendGateway is a helper to send messages over the gateway.
	SendGateway(ctx context.Context, m ws.Event) error
}

var (
	_ MainSession = (*session.Session)(nil)
	_ MainSession = (*state.State)(nil)
)

// Session is a single voice session that wraps around the voice gateway and UDP
// connection.
type Session struct {
	*handler.Handler
	session MainSession

	mut sync.RWMutex
	// connected is a non-nil blocking channel after Join is called and is
	// closed once Leave is called.
	disconnected chan struct{}

	state voicegateway.State // guarded except UserID

	detachReconnect []func()

	// udpManager is the manager for a UDP connection. The user can use this to
	// plug in a custom UDP dialer.
	udpManager *udp.Manager

	gateway  *voicegateway.Gateway
	dave     *davesession.Session
	gwCancel context.CancelFunc
	gwDone   <-chan struct{}

	// daveStatusMu protects a small diagnostic snapshot exposed to flow errors.
	daveStatusMu sync.Mutex
	daveLastEvent string
	daveEventCounts map[string]int
	daveLastSendError string
	// daveGuildID/daveChannelID/daveUsers track which channel the current DAVE
	// session belongs to, guarded by daveStatusMu, so voice-state events from
	// the main gateway can keep the recognized-member set current.
	daveGuildID   discord.GuildID
	daveChannelID discord.ChannelID
	daveLog       *daveLogBuffer

	WSTimeout      time.Duration // global WSTimeout
	WSMaxRetry     int           // 2
	WSRetryDelay   time.Duration // 2s
	WSWaitDuration time.Duration // 5s

	// joining determines the behavior of incoming event callbacks (Update).
	// If this is true, incoming events will just send into Updated channels. If
	// false, events will trigger a reconnection.
	joining moreatomic.Bool
	// disconnectClosed is true if connected is already closed. It is only used
	// to keep track of closing connected.
	disconnectClosed bool
}

// NewSession creates a new voice session for the current user.
func NewSession(state MainSession) (*Session, error) {
	u, err := state.Me()
	if err != nil {
		return nil, fmt.Errorf("failed to get me: %w", err)
	}

	return NewSessionCustom(state, u.ID), nil
}

// NewSessionCustom creates a new voice session from the given session and user
// ID.
func NewSessionCustom(ses MainSession, userID discord.UserID) *Session {
	closed := make(chan struct{})
	close(closed)

	session := &Session{
		Handler: handler.New(),
		session: ses,
		state: voicegateway.State{
			UserID: userID,
		},
		udpManager:     udp.NewManager(),
		WSTimeout:      WSTimeout,
		WSMaxRetry:     2,
		WSRetryDelay:   2 * time.Second,
		WSWaitDuration: 5 * time.Second,

		// Set this pair of value in so we never have to nil-check the channel.
		// We can just assume that it's either closed or connected.
		disconnected:     closed,
		disconnectClosed: true,
		daveEventCounts: make(map[string]int),
	}

	session.AddHandler(func(ev *ws.BinaryEvent) {
		if session.dave != nil {
			session.recordDaveEvent("binary")
			session.handleDaveBinary(ev)
		}
	})
	session.AddHandler(func(ev *voicegateway.ClientsConnectEvent) {
		if session.dave != nil {
			session.recordDaveEvent("clients_connect")
			for _, userID := range ev.UserIDs {
				session.dave.AddUser(daveUserID(userID))
			}
		}
	})
	session.AddHandler(func(ev *voicegateway.ClientConnectEvent) {
		if session.dave != nil {
			session.recordDaveEvent("client_connect")
			session.dave.AddUser(daveUserID(ev.UserID))
		}
	})
	session.AddHandler(func(ev *voicegateway.ClientDisconnectEvent) {
		if session.dave != nil {
			session.recordDaveEvent("client_disconnect")
			session.dave.RemoveUser(daveUserID(ev.UserID))
		}
	})
	session.AddHandler(func(ev *voicegateway.DaveProtocolPrepareTransitionEvent) {
		if session.dave != nil {
			session.recordDaveEvent("prepare_transition")
			session.dave.OnDavePrepareTransition(ev.TransitionID, ev.ProtocolVersion)
		}
	})
	session.AddHandler(func(ev *voicegateway.DaveProtocolExecuteTransitionEvent) {
		if session.dave != nil {
			session.recordDaveEvent("execute_transition")
			session.dave.OnDaveExecuteTransition(ev.TransitionID)
		}
	})
	session.AddHandler(func(ev *voicegateway.DaveProtocolPrepareEpochEvent) {
		if session.dave != nil {
			session.recordDaveEvent("prepare_epoch")
			session.dave.OnDavePrepareEpoch(ev.Epoch, ev.ProtocolVersion)
		}
	})
	return session
}

func (s *Session) recordDaveEvent(name string) {
	s.daveStatusMu.Lock()
	defer s.daveStatusMu.Unlock()
	if s.daveEventCounts == nil {
		s.daveEventCounts = make(map[string]int)
	}
	s.daveLastEvent = name
	s.daveEventCounts[name]++
}

// DaveDiagnostics returns safe, compact state for surfacing handshake failures
// in flow errors when application logs are not available.
func (s *Session) DaveDiagnostics() string {
	s.daveStatusMu.Lock()
	last := s.daveLastEvent
	lastSendError := s.daveLastSendError
	counts := make(map[string]int, len(s.daveEventCounts))
	for name, count := range s.daveEventCounts {
		counts[name] = count
	}
	s.daveStatusMu.Unlock()

	s.daveStatusMu.Lock()
	dave := s.dave
	logs := ""
	if s.daveLog != nil {
		logs = s.daveLog.String()
	}
	s.daveStatusMu.Unlock()

	if dave == nil {
		return fmt.Sprintf("no DAVE session; last gateway event=%q; event counts=%v; DAVE log: %s", last, counts, logs)
	}
	state := dave.State()
	stats := dave.Stats()
	return fmt.Sprintf("protocol version=%d, ready=%t, epoch=%d; last gateway event=%q; event counts=%v; MLS commits=%d processed/%d failed, welcomes=%d joined/%d failed, proposals rejected=%d; last send error=%q; DAVE log: %s",
		state.ProtocolVersion, state.Ready, state.EpochID, last, counts,
		stats.CommitsProcessed, stats.CommitsFailed, stats.WelcomesJoined, stats.WelcomesFailed,
		stats.ProposalsRejected, lastSendError, logs)
}

type daveCallbacks struct{ session *Session }

func (c daveCallbacks) SendMLSKeyPackage(data []byte) error {
	return c.session.sendDaveBinary(26, data)
}
func (c daveCallbacks) SendMLSCommitWelcome(data []byte) error {
	return c.session.sendDaveBinary(28, data)
}
func (c daveCallbacks) SendReadyForTransition(id uint16) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.session.gateway.Send(ctx, &voicegateway.DaveProtocolReadyForTransitionCommand{TransitionID: id})
}
func (c daveCallbacks) SendInvalidCommitWelcome(id uint16) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.session.gateway.Send(ctx, &voicegateway.DaveMLSInvalidCommitWelcomeCommand{TransitionID: id})
}

func (s *Session) sendDaveBinary(op byte, payload []byte) error {
	data := make([]byte, 1+len(payload))
	data[0] = op
	copy(data[1:], payload)

	// The voice gateway can still be settling right after Session Description
	// (which is exactly when the key package goes out). The DAVE library never
	// re-sends a key package it thinks it already sent, so retry here instead
	// of losing it.
	var err error
	for attempt := 0; attempt < 5; attempt++ {
		if attempt > 0 {
			time.Sleep(300 * time.Millisecond)
		}
		gw := s.gateway
		if gw == nil {
			err = errors.New("voice gateway not ready")
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = gw.SendBinary(ctx, data)
		cancel()
		if err == nil {
			break
		}
	}
	if err != nil {
		s.daveStatusMu.Lock()
		s.daveLastSendError = fmt.Sprintf("binary opcode %d: %v", op, err)
		s.daveStatusMu.Unlock()
		return err
	}
	s.recordDaveEvent(fmt.Sprintf("sent_binary_opcode_%d", op))
	return nil
}

func (s *Session) handleDaveBinary(ev *ws.BinaryEvent) {
	if len(ev.Data) == 0 || s.dave == nil {
		s.recordDaveEvent("malformed_binary")
		return
	}
	// Voice gateway v8 prefixes server DAVE frames with a 2-byte sequence
	// number. Older voice gateway versions send opcode-first frames. Prefer
	// the v8 layout, with an opcode-first fallback for older servers.
	op, payload := byte(0), []byte(nil)
	if len(ev.Data) >= 3 && isServerDaveOpcode(ev.Data[2]) {
		if s.gateway != nil {
			s.gateway.AckSequence(int64(binary.BigEndian.Uint16(ev.Data[:2])))
		}
		op, payload = ev.Data[2], ev.Data[3:]
	} else if isServerDaveOpcode(ev.Data[0]) {
		op, payload = ev.Data[0], ev.Data[1:]
	} else {
		s.recordDaveEvent(fmt.Sprintf("unknown_binary_opcode_%d", ev.Data[0]))
		return
	}
	s.recordDaveEvent(fmt.Sprintf("binary_opcode_%d", op))
	if op == 27 && len(payload) > 0 {
		s.recordDaveEvent(fmt.Sprintf("binary_opcode_27_operation_%d", payload[0]))
	}
	switch op {
	case 25:
		s.dave.OnDaveMLSExternalSenderPackage(payload)
	case 27:
		s.dave.OnDaveMLSProposals(payload)
	case 29:
		if len(payload) < 2 {
			return
		}
		id := binary.BigEndian.Uint16(payload[:2])
		s.dave.OnDaveMLSPrepareCommitTransition(id, payload[2:])
	case 30:
		if len(payload) < 2 {
			return
		}
		id := binary.BigEndian.Uint16(payload[:2])
		s.dave.OnDaveMLSWelcome(id, payload[2:])
	}
}

func isServerDaveOpcode(op byte) bool {
	switch op {
	case 25, 27, 29, 30:
		return true
	default:
		return false
	}
}

// SetUDPDialer sets the given dialer to be used for dialing UDP voice
// connections.
func (s *Session) SetUDPDialer(d udp.DialFunc) {
	s.udpManager.SetDialer(d)
}

func (s *Session) acquireUpdate(f func()) bool {
	if s.joining.Get() {
		return false
	}

	s.mut.Lock()
	defer s.mut.Unlock()

	// Ignore if we haven't connected yet or we're still joining.
	if s.udpManager.IsClosed() {
		return false
	}

	f()
	return true
}

// updateServer is specifically used to monitor for reconnects.
func (s *Session) updateServer(ev *gateway.VoiceServerUpdateEvent) {
	s.acquireUpdate(func() {
		if s.state.GuildID != ev.GuildID {
			return
		}

		s.state.Endpoint = ev.Endpoint
		s.state.Token = ev.Token

		ctx, cancel := context.WithTimeout(context.Background(), WSTimeout)
		defer cancel()

		s.reconnectCtx(ctx)
	})
}

// updateState is specifically used after connecting to monitor when the bot is
// forced across channels.
func (s *Session) updateState(ev *gateway.VoiceStateUpdateEvent) {
	s.acquireUpdate(func() {
		if s.state.GuildID != ev.GuildID || s.state.UserID != ev.UserID {
			return
		}

		s.state.ChannelID = ev.ChannelID
		s.state.SessionID = ev.SessionID

		ctx, cancel := context.WithTimeout(context.Background(), WSTimeout)
		defer cancel()

		s.reconnectCtx(ctx)
	})
}

// JoinChannelAndSpeak is a convenient function that calls JoinChannel then
// Speaking.
func (s *Session) JoinChannelAndSpeak(ctx context.Context, chID discord.ChannelID, mute, deaf bool) error {
	if err := s.JoinChannel(ctx, chID, mute, deaf); err != nil {
		return fmt.Errorf("cannot join channel: %w", err)
	}
	return s.Speaking(ctx, voicegateway.Microphone)
}

type waitEventChs struct {
	serverUpdate chan *gateway.VoiceServerUpdateEvent
	stateUpdate  chan *gateway.VoiceStateUpdateEvent
}

// JoinChannel joins the given voice channel with the default timeout.
func (s *Session) JoinChannel(ctx context.Context, chID discord.ChannelID, mute, deaf bool) error {
	var ch *discord.Channel

	if chID.IsValid() {
		var err error
		ch, err = s.session.Channel(chID)
		if err != nil {
			return fmt.Errorf("invalid channel ID: %w", err)
		}
	}

	s.mut.Lock()
	defer s.mut.Unlock()

	// Error out if we're already joining. JoinChannel shouldn't be called
	// concurrently.
	if !s.joining.Acquire() {
		return errors.New("JoinChannel working elsewhere")
	}

	defer s.joining.Set(false)

	// Set the state.
	if ch != nil {
		s.state.ChannelID = ch.ID
		s.state.GuildID = ch.GuildID
	} else {
		s.state.GuildID = 0
		// Ensure that if `cID` is zero that it passes null to the update event.
		s.state.ChannelID = discord.NullChannelID
	}

	if s.detachReconnect == nil {
		s.detachReconnect = []func(){
			s.session.AddHandler(s.updateServer),
			s.session.AddHandler(s.updateState),
			s.session.AddHandler(s.trackDaveMember),
		}
	}

	chs := waitEventChs{
		serverUpdate: make(chan *gateway.VoiceServerUpdateEvent),
		stateUpdate:  make(chan *gateway.VoiceStateUpdateEvent),
	}

	// Bind the handlers.
	cancels := []func(){
		s.session.AddHandler(chs.serverUpdate),
		s.session.AddHandler(chs.stateUpdate),
	}
	// Disconnects the handlers once the function exits.
	defer func() {
		for _, cancel := range cancels {
			cancel()
		}
	}()

	// Ensure gateway and voiceUDP are already closed.
	s.ensureClosed()

	// https://discord.com/developers/docs/topics/voice-connections#retrieving-voice-server-information
	// Send a Voice State Update event to the gateway.
	data := &gateway.UpdateVoiceStateCommand{
		GuildID:   s.state.GuildID,
		ChannelID: s.state.ChannelID,
		SelfMute:  mute,
		SelfDeaf:  deaf,
	}

	var err error
	var timer *time.Timer

	// Retry 3 times maximum.
	for i := 0; i < s.WSMaxRetry; i++ {
		if err = s.askDiscord(ctx, data, chs); err == nil {
			break
		}

		// If this is the first attempt and the context timed out, it's
		// probably the context that's waiting for gateway events. Retry the
		// loop.
		if i == 0 && errors.Is(err, ctx.Err()) {
			continue
		}

		if timer == nil {
			// Set up a timer.
			timer = time.NewTimer(s.WSRetryDelay)
			defer timer.Stop()
		} else {
			timer.Reset(s.WSRetryDelay)
		}

		select {
		case <-timer.C:
			continue
		case <-ctx.Done():
			return fmt.Errorf("cannot ask Discord for events: %w", err)
		}
	}

	// These 2 methods should've updated s.state before sending into these
	// channels. Since s.state is already filled, we can go ahead and connect.

	// Mark the session as connected and move on. This allows one of the
	// connected handlers to reconnect on its own.
	s.disconnected = make(chan struct{})

	return s.reconnectCtx(ctx)
}

func (s *Session) askDiscord(
	ctx context.Context, data *gateway.UpdateVoiceStateCommand, chs waitEventChs) error {

	// https://discord.com/developers/docs/topics/voice-connections#retrieving-voice-server-information
	// Send a Voice State Update event to the gateway.
	if err := s.session.SendGateway(ctx, data); err != nil {
		return fmt.Errorf("failed to send Voice State Update event: %w", err)
	}

	// Wait for 2 replies. The above command should reply with these 2 events.
	if err := s.waitForIncoming(ctx, chs); err != nil {
		return fmt.Errorf("failed to wait for needed gateway events: %w", err)
	}

	return nil
}

func (s *Session) waitForIncoming(ctx context.Context, chs waitEventChs) error {
	ctx, cancel := context.WithTimeout(ctx, s.WSWaitDuration)
	defer cancel()

	state := false
	// server is true when we already have the token and endpoint, meaning that
	// we don't have to wait for another such event.
	server := s.state.Token != "" && s.state.Endpoint != ""

	// Loop until timeout or until we have all the information that we need.
	for !(server && state) {
		select {
		case ev := <-chs.serverUpdate:
			if s.state.GuildID != ev.GuildID {
				continue
			}
			s.state.Endpoint = ev.Endpoint
			s.state.Token = ev.Token
			server = true

		case ev := <-chs.stateUpdate:
			if s.state.GuildID != ev.GuildID || s.state.UserID != ev.UserID {
				continue
			}
			s.state.SessionID = ev.SessionID
			s.state.ChannelID = ev.ChannelID
			state = true

		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

// reconnect uses the current state to reconnect to a new gateway and UDP
// connection.
func (s *Session) reconnectCtx(ctx context.Context) error {
	ws.WSDebug("Sending stop handle.")

	if err := s.udpManager.Pause(ctx); err != nil {
		return fmt.Errorf("cannot pause UDP manager: %w", err)
	}
	defer func() {
		if !s.udpManager.Continue() {
			panic("UDP manager continued but invalid lock ownership")
		}
	}()

	s.ensureClosed()

	ws.WSDebug("Start gateway.")
	s.gateway = voicegateway.New(s.state)
	s.daveStatusMu.Lock()
	if s.daveLog == nil {
		// Kept across reconnects so earlier handshake attempts stay visible.
		s.daveLog = newDaveLogBuffer(40)
	}
	logBuf := s.daveLog
	s.daveStatusMu.Unlock()
	daveSess := davesession.New(
		daveUserID(s.state.UserID),
		daveCallbacks{session: s},
		davesession.WithLogger(slog.New(slog.NewTextHandler(logBuf, &slog.HandlerOptions{Level: slog.LevelInfo}))),
	)
	s.daveStatusMu.Lock()
	s.dave = daveSess
	s.daveGuildID = s.state.GuildID
	s.daveChannelID = s.state.ChannelID
	s.daveStatusMu.Unlock()

	// The bot itself is always a member of its voice/DAVE session.
	s.dave.AddUser(daveUserID(s.state.UserID))
	s.dave.SetChannelID(godave.ChannelID(s.state.ChannelID))
	// The DAVE library silently ignores "add member" proposals for users it
	// was not told about, and Discord does not reliably send CLIENTS_CONNECT
	// for people already in the channel. Seed the recognized-member set from
	// the main gateway's voice states so the MLS commit is not skipped.
	s.seedDaveUsers()

	// Open the voice gateway. The function will block until Ready is received.
	gwctx, gwcancel := context.WithCancel(context.Background())
	s.gwCancel = gwcancel

	gwch := s.gateway.Connect(gwctx)
	ws.WSDebug("Voice Gateway connected")

	if err := s.spinGateway(ctx, gwch); err != nil {
		ws.WSDebug("Voice spinGateway error:", err)
		// Early cancel the gateway.
		gwcancel()
		// Nil this so future reconnects don't use the invalid gwDone.
		s.gwCancel = nil
		// Emit the error. It's fine to do this here since this is the only
		// place that can error out.
		s.Handler.Call(&ReconnectError{err})
		return fmt.Errorf("cannot wait for event sequence from voice gateway: %w", err)
	}

	// Start dispatching.
	s.gwDone = ophandler.Loop(gwch, s.Handler)

	ws.WSDebug("Voice reconnectCtx finished with no error")

	return nil
}

func (s *Session) spinGateway(ctx context.Context, gwch <-chan ws.Op) error {
	var err error
	var conn *udp.Connection

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ev, ok := <-gwch:
			if !ok {
				return fmt.Errorf("voice gateway error: %w", s.gateway.LastError())
			}

			switch data := ev.Data.(type) {
			case *ws.CloseEvent:
				return fmt.Errorf("voice gateway error: %w", data)

			case *voicegateway.ReadyEvent:
				ws.WSDebug("Got ready from voice gateway, SSRC:", data.SSRC)
				s.dave.AssignSsrcToCodec(data.SSRC, godave.CodecOpus)

				if !data.SupportsMode(Protocol) {
					return fmt.Errorf("voice server does not advertise required encryption mode %q; modes=%v", Protocol, data.Modes)
				}

				// Prepare the UDP voice connection.
				conn, err = s.udpManager.Dial(ctx, data.Addr(), data.SSRC)
				if err != nil {
					return fmt.Errorf("failed to open voice UDP connection: %w", err)
				}

				if err := s.gateway.Send(ctx, &voicegateway.SelectProtocolCommand{
					Protocol: "udp",
					Data: voicegateway.SelectProtocolData{
						Address: conn.GatewayIP,
						Port:    conn.GatewayPort,
						Mode:    Protocol,
					},
				}); err != nil {
					return fmt.Errorf("failed to send SelectProtocolCommand: %w", err)
				}

			case *voicegateway.SessionDescriptionEvent:
				if conn == nil {
					return errors.New("server bug: SessionDescription before Ready")
				}

				ws.WSDebug("Received secret key from voice gateway")

				// We're done.
				if data.Mode != Protocol {
					return fmt.Errorf("voice server selected unsupported encryption mode %q; expected %q", data.Mode, Protocol)
				}

				conn.UseMode(data.Mode)
				conn.UseSecret(data.SecretKey)
				conn.SetDaveSession(s.dave)
				s.udpManager.SetDaveSession(s.dave)
				s.dave.OnSelectProtocolAck(data.DaveProtocolVersion)
				return nil
			}

			// Dispatch this event to the handler.
			s.Handler.Call(ev.Data)
		}
	}
}

// WaitDaveReady waits for the DAVE MLS epoch when the voice gateway negotiated DAVE.
func (s *Session) WaitDaveReady(ctx context.Context) error {
	if s.dave == nil {
		return nil
	}
	if s.dave.State().ProtocolVersion == 0 {
		return nil
	}
	if _, err := s.dave.WaitReady(ctx); err != nil {
		return fmt.Errorf("timed out waiting for Discord DAVE encryption: %w", err)
	}
	return nil
}

// Speaking tells Discord we're speaking. This method should not be called
// concurrently.
//
// If only NotSpeaking (0) is given, then even if the gateway cannot be reached,
// a nil error will be returned. This is because sending Discord a not-speaking
// event is a destruction command that doesn't affect the outcome of anything
// done after whatsoever.
func (s *Session) Speaking(ctx context.Context, flag voicegateway.SpeakingFlag) error {
	s.mut.Lock()
	gateway := s.gateway
	s.mut.Unlock()

	if err := gateway.Speaking(ctx, flag); err != nil && flag != 0 {
		return err
	}

	return nil
}

// Write writes into the UDP voice connection. This method is thread safe as far
// as calling other methods of Session goes; HOWEVER it is not thread safe to
// call Write itself concurrently.
func (s *Session) Write(b []byte) (int, error) {
	return s.udpManager.Write(b)
}

// ReadPacket reads a single packet from the UDP connection. This is NOT at all
// thread safe, and must be used very carefully. The backing buffer is always
// reused.
func (s *Session) ReadPacket() (*udp.Packet, error) {
	return s.udpManager.ReadPacket()
}

// Leave disconnects the current voice session from the currently connected
// channel.
func (s *Session) Leave(ctx context.Context) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	s.ensureClosed()

	// Unbind the handlers.
	if s.detachReconnect != nil {
		for _, detach := range s.detachReconnect {
			detach()
		}
		s.detachReconnect = nil
	}

	// If we're already closed.
	if s.gateway == nil && s.udpManager.IsClosed() {
		return nil
	}

	// Notify Discord that we're leaving.
	sendErr := s.session.SendGateway(ctx, &gateway.UpdateVoiceStateCommand{
		GuildID:   s.state.GuildID,
		ChannelID: discord.ChannelID(discord.NullSnowflake),
		SelfMute:  true,
		SelfDeaf:  true,
	})

	// Wait for the gateway to exit first before we tell the user of the gateway
	// send error.
	if err := s.cancelGateway(ctx); err != nil {
		return err
	}

	if sendErr != nil {
		return fmt.Errorf("failed to update voice state: %w", sendErr)
	}

	return nil
}

func (s *Session) cancelGateway(ctx context.Context) error {
	if s.gwCancel != nil {
		s.gwCancel()
		s.gwCancel = nil

		// Wait for the previous gateway to finish closing up, but make sure to
		// bail if the context expires.
		if err := ophandler.WaitForDone(ctx, s.gwDone); err != nil {
			return fmt.Errorf("cannot wait for gateway to close: %w", err)
		}
	}

	return nil
}

const (
	permanentClose = true
	temporaryClose = false
)

// close ensures everything is closed. It does not acquire the mutex.
func (s *Session) ensureClosed() {
	s.daveStatusMu.Lock()
	oldDave := s.dave
	s.dave = nil
	s.daveStatusMu.Unlock()
	if oldDave != nil {
		_ = oldDave.Close()
	}
	// Disconnect the UDP connection. If not permanent, then pause.
	s.udpManager.Close()

	if s.gwCancel != nil {
		s.gwCancel()
		// Don't actually clear this field, because we still want the caller to
		// be able to wait for the gateway to completely exit using
		// cancelGateway.
	}
}


// daveUserID converts a Discord user ID into the identifier the DAVE library
// expects: the decimal string of the snowflake. A plain godave.UserID(id)
// conversion of a uint64 yields a single rune instead, which breaks the MLS
// credential identity and every recognized-member check.
func daveUserID(id discord.UserID) godave.UserID {
	return godave.UserID(strconv.FormatUint(uint64(id), 10))
}

// daveLogBuffer keeps the most recent lines logged by the DAVE library so
// handshake failures can be reported without access to server logs.
type daveLogBuffer struct {
	mu    sync.Mutex
	max   int
	lines []string
}

func newDaveLogBuffer(max int) *daveLogBuffer { return &daveLogBuffer{max: max} }

func (b *daveLogBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, line := range strings.Split(strings.TrimRight(string(p), "\n"), "\n") {
		if line == "" {
			continue
		}
		if len(line) > 220 {
			line = line[:220] + "..."
		}
		b.lines = append(b.lines, line)
	}
	if len(b.lines) > b.max {
		b.lines = b.lines[len(b.lines)-b.max:]
	}
	return len(p), nil
}

func (b *daveLogBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.lines) == 0 {
		return "(none)"
	}
	return strings.Join(b.lines, " | ")
}

// voiceStateLister is implemented by *state.State (through its store
// Cabinet) and gives access to the voice states the main gateway has cached.
type voiceStateLister interface {
	VoiceStates(discord.GuildID) ([]discord.VoiceState, error)
}

// seedDaveUsers registers everyone currently in the voice channel as a
// recognized DAVE member.
func (s *Session) seedDaveUsers() {
	lister, ok := s.session.(voiceStateLister)
	if !ok {
		return
	}
	s.daveStatusMu.Lock()
	dave, guildID, channelID := s.dave, s.daveGuildID, s.daveChannelID
	s.daveStatusMu.Unlock()
	if dave == nil || !guildID.IsValid() {
		return
	}
	states, err := lister.VoiceStates(guildID)
	if err != nil {
		return
	}
	n := 0
	for _, vs := range states {
		if vs.ChannelID == channelID {
			dave.AddUser(daveUserID(vs.UserID))
			n++
		}
	}
	s.recordDaveEvent(fmt.Sprintf("seeded_%d_channel_members", n))
}

// trackDaveMember keeps the DAVE recognized-member set in sync with people
// joining and leaving the bot's voice channel after the initial seed.
func (s *Session) trackDaveMember(ev *gateway.VoiceStateUpdateEvent) {
	s.daveStatusMu.Lock()
	dave, guildID, channelID := s.dave, s.daveGuildID, s.daveChannelID
	s.daveStatusMu.Unlock()
	if dave == nil || ev.GuildID != guildID || ev.UserID == s.state.UserID {
		return
	}
	if ev.ChannelID == channelID {
		dave.AddUser(daveUserID(ev.UserID))
	} else {
		dave.RemoveUser(daveUserID(ev.UserID))
	}
}

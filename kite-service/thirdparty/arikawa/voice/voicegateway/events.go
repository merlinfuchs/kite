package voicegateway

import (
	"strconv"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/ws"
)

//go:generate go run ../../utils/cmd/genevent -p voicegateway -o event_methods.go

// OpUnmarshalers contains the Op unmarshalers for the voice gateway events.
var OpUnmarshalers = ws.NewOpUnmarshalers()

// IdentifyCommand is a command for Op 0.
//
// https://discord.com/developers/docs/topics/voice-connections#establishing-a-voice-websocket-connection-example-voice-identify-payload
type IdentifyCommand struct {
	GuildID   discord.GuildID `json:"server_id"` // yes, this should be "server_id"
	UserID    discord.UserID  `json:"user_id"`
	SessionID string          `json:"session_id"`
	Token     string          `json:"token"`
	// MaxDaveProtocolVersion declares the highest DAVE (E2EE) protocol
	// version this client supports. Discord requires this to be >= 1 for
	// non-stage voice connections since March 2026, or it closes the
	// connection with code 4017 ("E2EE/DAVE protocol required"). This
	// client does not implement DAVE's MLS encryption; declaring version 1
	// without completing the handshake relies on Discord's transitional
	// passthrough behavior and is not guaranteed to keep working.
	MaxDaveProtocolVersion int `json:"max_dave_protocol_version"`
}

// SelectProtocolCommand is a command for Op 1.
//
// https://discord.com/developers/docs/topics/voice-connections#establishing-a-voice-udp-connection-example-select-protocol-payload
type SelectProtocolCommand struct {
	Protocol string             `json:"protocol"`
	Data     SelectProtocolData `json:"data"`
}

// SelectProtocolData is the data inside a SelectProtocolCommand.
type SelectProtocolData struct {
	Address string `json:"address"`
	Port    uint16 `json:"port"`
	Mode    string `json:"mode"`
}

// ReadyEvent is an event for Op 2.
//
// https://discord.com/developers/docs/topics/voice-connections#establishing-a-voice-websocket-connection-example-voice-ready-payload
type ReadyEvent struct {
	SSRC        uint32   `json:"ssrc"`
	IP          string   `json:"ip"`
	Port        int      `json:"port"`
	Modes       []string `json:"modes"`
	Experiments []string `json:"experiments"`

	// From Discord's API Docs:
	//
	// `heartbeat_interval` here is an erroneous field and should be ignored.
	// The correct `heartbeat_interval` value comes from the Hello payload.

	// HeartbeatInterval discord.Milliseconds `json:"heartbeat_interval"`
}

// Addr formats the URL inside Ready to be of format "host:port".
func (r ReadyEvent) Addr() string {
	return r.IP + ":" + strconv.Itoa(r.Port)
}

// SupportsMode reports whether the voice server advertised the given transport
// encryption mode in its Ready event.
func (r *ReadyEvent) SupportsMode(mode string) bool {
	for _, m := range r.Modes {
		if m == mode {
			return true
		}
	}
	return false
}

// HeartbeatCommand is a command for Op 3 on voice gateway v8.
type HeartbeatCommand struct {
	Time   uint64 `json:"t"`
	SeqAck int64  `json:"seq_ack"`
}

// SessionDescriptionEvent is an event for Op 4.
//
// https://discord.com/developers/docs/topics/voice-connections#establishing-a-voice-udp-connection-example-session-description-payload
type SessionDescriptionEvent struct {
	Mode                string   `json:"mode"`
	SecretKey           [32]byte `json:"secret_key"`
	DaveProtocolVersion uint16   `json:"dave_protocol_version"`
}

type DaveProtocolPrepareTransitionEvent struct {
	TransitionID    uint16 `json:"transition_id"`
	ProtocolVersion uint16 `json:"protocol_version"`
}

type DaveProtocolExecuteTransitionEvent struct {
	TransitionID uint16 `json:"transition_id"`
}

type DaveProtocolPrepareEpochEvent struct {
	Epoch           int    `json:"epoch"`
	ProtocolVersion uint16 `json:"protocol_version"`
}

type DaveProtocolReadyForTransitionCommand struct {
	TransitionID uint16 `json:"transition_id"`
}

type DaveMLSInvalidCommitWelcomeCommand struct {
	TransitionID uint16 `json:"transition_id"`
}

// https://discord.com/developers/docs/topics/voice-connections#speaking
type SpeakingFlag uint64

const NotSpeaking SpeakingFlag = 0

const (
	Microphone SpeakingFlag = 1 << iota
	Soundshare
	Priority
)

// SpeakingEvent is an event for Op 5. It is also a command.
//
// https://discord.com/developers/docs/topics/voice-connections#speaking-example-speaking-payload
type SpeakingEvent struct {
	Speaking SpeakingFlag   `json:"speaking"`
	Delay    int            `json:"delay"`
	SSRC     uint32         `json:"ssrc"`
	UserID   discord.UserID `json:"user_id,omitempty"`
}

// HeartbeatAckEvent is an event for Op 6.
//
// https://discord.com/developers/docs/topics/voice-connections#heartbeating-example-heartbeat-ack-payload
type HeartbeatAckEvent struct {
	Time uint64 `json:"t"`
}

// ResumeCommand is a command for Op 7.
//
// https://discord.com/developers/docs/topics/voice-connections#resuming-voice-connection-example-resume-connection-payload
type ResumeCommand struct {
	GuildID   discord.GuildID `json:"server_id"` // yes, this should be "server_id"
	SessionID string          `json:"session_id"`
	Token     string          `json:"token"`
	SeqAck    int64           `json:"seq_ack"`
}

// HelloEvent is an event for Op 8.
//
// https://discord.com/developers/docs/topics/voice-connections#heartbeating-example-hello-payload-since-v3
type HelloEvent struct {
	HeartbeatInterval discord.Milliseconds `json:"heartbeat_interval"`
}

// ResumedEvent is an event for Op 9.
// https://discord.com/developers/docs/topics/voice-connections#resuming-voice-connection-example-resumed-payload
type ResumedEvent struct{}

// ClientsConnectEvent is the DAVE participant-list event for Op 11.
// Discord uses this event to tell DAVE clients which user IDs are currently
// expected in the media session. The older CLIENT_CONNECT (Op 12) event is
// retained below for compatibility with the legacy voice event stream.
type ClientsConnectEvent struct {
	UserIDs []discord.UserID `json:"user_ids"`
}

// ClientConnectEvent is an event for Op 12. It is undocumented.
type ClientConnectEvent struct {
	UserID    discord.UserID `json:"user_id"`
	AudioSSRC uint32         `json:"audio_ssrc"`
	VideoSSRC uint32         `json:"video_ssrc"`
}

// ClientDisconnectEvent is an event for Op 13. It is undocumented, but its
// existence is mentioned in this issue:
// https://github.com/discord/discord-api-docs/issues/510.
type ClientDisconnectEvent struct {
	UserID discord.UserID `json:"user_id"`
}

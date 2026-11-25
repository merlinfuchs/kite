package flow

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
)

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeActionTTSSpeak: executeActionTTSSpeak,
	})
}

// ttsDefaultVoice is the voice of a block without one.
const ttsDefaultVoice = "alloy"

// executeActionTTSSpeak turns the text into speech and plays it in a voice
// channel, like Play audio does with an uploaded sound.
func executeActionTTSSpeak(n *CompiledFlowNode, ctx *FlowContext) error {
	data := n.Data.TTSData
	if data == nil {
		return &FlowError{
			Code:    FlowNodeErrorUnknown,
			Message: "tts_data is nil",
		}
	}
	if ctx.TTS == nil {
		return provider.ErrTTSUnavailable
	}

	text, err := ctx.EvalTemplate(data.Text)
	if err != nil {
		return err
	}
	speech := strings.TrimSpace(text.String())
	if speech == "" {
		return errors.New("text is empty")
	}
	if length := utf8.RuneCountInString(speech); length > ttsMaxLength {
		return fmt.Errorf("text is %d characters, more than the limit of %d", length, ttsMaxLength)
	}

	voice := data.Voice
	if voice == "" {
		voice = ttsDefaultVoice
	}
	if _, ok := ttsVoices[voice]; !ok {
		return fmt.Errorf("unsupported voice %q", voice)
	}
	if data.Speed != 0 && (data.Speed < ttsMinSpeed || data.Speed > ttsMaxSpeed) {
		return fmt.Errorf("speed must be between %g and %g", ttsMinSpeed, ttsMaxSpeed)
	}

	if n.Data.ChannelTarget == "" {
		return errors.New("set the voice channel to speak in")
	}
	channelTarget, err := ctx.EvalTemplate(n.Data.ChannelTarget)
	if err != nil {
		return err
	}

	// Like Play audio, the guild comes from the channel. All of this is
	// checked before the speech is made, which costs money.
	channel, err := ctx.Discord.Channel(ctx, discord.ChannelID(channelTarget.Snowflake()))
	if err != nil {
		return err
	}
	if !channelTypeSupportsVoice(channel.Type) {
		return fmt.Errorf("channel %s is not a voice channel", channel.ID)
	}
	if channel.Type == discord.GuildStageVoice {
		// The bot joins stages as a listener, nobody would hear it.
		return fmt.Errorf("speaking in stage channels isn't supported")
	}

	audio, err := ctx.TTS.Speak(ctx, provider.SpeakOpts{
		Text:  speech,
		Voice: voice,
		Speed: data.Speed,
	})
	if err != nil {
		return err
	}
	if len(audio.Content) == 0 {
		return errors.New("the speech is empty")
	}

	err = ctx.Discord.PlayAudio(ctx, provider.PlayAudioOpts{
		GuildID:    channel.GuildID,
		ChannelID:  channel.ID,
		Sound:      audio.Content,
		LeaveAfter: n.Data.AudioLeaveAfter,
	})
	if err != nil {
		return err
	}

	return n.ExecuteChildren(ctx)
}

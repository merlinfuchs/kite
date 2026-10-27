package notify

import (
	"context"
	"log/slog"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
)

const sendTimeout = 10 * time.Second

// Notifier sends Kite's own messages with the bot of the Kite server: direct
// messages to users, and posts for the admins in the review channel. It does
// nothing without a bot token, and a nil Notifier does nothing either.
type Notifier struct {
	client          *api.Client
	reviewChannelID discord.ChannelID
	// appBaseURL is where the app runs, for links in the messages.
	appBaseURL string
}

func NewNotifier(botToken string, reviewChannelID string, appBaseURL string) *Notifier {
	if botToken == "" {
		return nil
	}
	n := &Notifier{client: api.NewClient("Bot " + botToken), appBaseURL: appBaseURL}
	if id, err := discord.ParseSnowflake(reviewChannelID); err == nil {
		n.reviewChannelID = discord.ChannelID(id)
	}
	return n
}

// DM sends the content to the user with the Discord ID in the background.
// Discord only delivers it if the user shares a server with the bot and
// allows direct messages from it.
func (n *Notifier) DM(discordUserID string, content string) {
	if n == nil {
		return
	}
	id, err := discord.ParseSnowflake(discordUserID)
	if err != nil {
		return
	}
	go n.send("dm", func(c *api.Client) error {
		ch, err := c.CreatePrivateChannel(discord.UserID(id))
		if err != nil {
			return err
		}
		return sendMessage(c, ch.ID, content)
	})
}

// Review posts the content in the review channel in the background, if
// there is one.
func (n *Notifier) Review(content string) {
	if n == nil || !n.reviewChannelID.IsValid() {
		return
	}
	go n.send("review", func(c *api.Client) error {
		return sendMessage(c, n.reviewChannelID, content)
	})
}

func (n *Notifier) send(kind string, f func(c *api.Client) error) {
	ctx, cancel := context.WithTimeout(context.Background(), sendTimeout)
	defer cancel()
	if err := f(n.client.WithContext(ctx)); err != nil {
		slog.Warn("Failed to send notification", slog.String("kind", kind), slog.String("error", err.Error()))
	}
}

// sendMessage sends the content without pinging anyone, since it has names
// users chose.
func sendMessage(c *api.Client, channelID discord.ChannelID, content string) error {
	_, err := c.SendMessageComplex(channelID, api.SendMessageData{
		Content:         content,
		AllowedMentions: &api.AllowedMentions{Parse: []api.AllowedMentionType{}},
	})
	return err
}

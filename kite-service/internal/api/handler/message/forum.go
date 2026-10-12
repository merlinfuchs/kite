package message

import (
	"mime/multipart"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/sendpart"
)

// forumPostData is the body of Discord's start thread in forum or media
// channel endpoint, which arikawa doesn't have.
// https://discord.com/developers/docs/resources/channel#start-thread-in-forum-or-media-channel
type forumPostData struct {
	Name        string              `json:"name"`
	AppliedTags []discord.TagID     `json:"applied_tags,omitempty"`
	Message     api.SendMessageData `json:"message"`
}

func (d forumPostData) NeedsMultipart() bool {
	return len(d.Message.Files) > 0
}

func (d forumPostData) WriteMultipart(body *multipart.Writer) error {
	return sendpart.Write(body, d, d.Message.Files)
}

// createForumPost creates a post in a forum or media channel with the message
// as its first message.
func createForumPost(client *api.Client, channelID discord.ChannelID, data forumPostData) (*discord.Channel, error) {
	var thread *discord.Channel
	err := sendpart.POST(client.Client, data, &thread, api.EndpointChannels+channelID.String()+"/threads")
	return thread, err
}

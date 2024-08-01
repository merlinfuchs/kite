package template

import (
	"fmt"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	statestore "github.com/diamondburned/arikawa/v3/state/store"
)

var standardDataMap = map[string]interface{}{}

type InteractionData struct {
	state *statestore.Cabinet
	i     *discord.InteractionEvent
}

func NewInteractionData(state *statestore.Cabinet, i *discord.InteractionEvent) *InteractionData {
	return &InteractionData{
		state: state,
		i:     i,
	}
}

func (d *InteractionData) User() interface{} {
	if d.i.Member != nil {
		res := NewMemberData(d.state, d.i.GuildID, d.i.Member)
		return &res
	}

	return NewUserData(d.i.User)
}

func (d *InteractionData) Member() *MemberData {
	if d.i.Member == nil {
		return nil
	}

	return NewMemberData(d.state, d.i.GuildID, d.i.Member)
}

func (d *InteractionData) Command() *CommandData {
	if d.i.Data.InteractionType() != discord.CommandInteractionType {
		return nil
	}

	data, _ := d.i.Data.(*discord.CommandInteraction)
	return NewCommandData(d.state, d.i.GuildID, data)
}

type UserData struct {
	u *discord.User
}

func NewUserData(u *discord.User) *UserData {
	return &UserData{u: u}
}

func (d *UserData) String() string {
	return d.Mention()
}

func (d *UserData) ID() string {
	return d.u.ID.String()
}

func (d *UserData) Name() string {
	if d.u.DisplayName != "" {
		return d.u.DisplayName
	}

	return d.u.Username
}

func (d *UserData) Username() string {
	return d.u.Username
}

func (d *UserData) GlobalName() string {
	return d.u.DisplayName
}

func (d *UserData) Discriminator() string {
	return d.u.Discriminator
}

func (d *UserData) Avatar() string {
	return d.u.Avatar
}

func (d *UserData) Banner() string {
	return d.u.Banner
}

func (d *UserData) Mention() string {
	return d.u.Mention()
}

func (d *UserData) AvatarURL() string {
	return d.u.AvatarURL()
}

func (d *UserData) BannerURL() string {
	return d.u.BannerURL()
}

type MemberData struct {
	UserData
	state   *statestore.Cabinet
	guildID discord.GuildID
	m       *discord.Member
}

func NewMemberData(state *statestore.Cabinet, guildID discord.GuildID, m *discord.Member) *MemberData {
	return &MemberData{
		UserData: UserData{&m.User},
		state:    state,
		guildID:  guildID,
		m:        m,
	}
}

func (d *MemberData) Nick() string {
	return d.m.Nick
}

func (d *MemberData) Roles() []*RoleData {
	res := make([]*RoleData, len(d.m.RoleIDs))
	for i, roleID := range d.m.RoleIDs {
		res[i] = NewRoleData(d.state, d.guildID, roleID, nil)
	}

	return res
}

func (d *MemberData) JoinedAt() time.Time {
	return d.m.Joined.Time()
}

func (d *MemberData) Name() string {
	if d.m.Nick != "" {
		return d.m.Nick
	}

	return d.UserData.Name()
}

func (d *MemberData) Avatar() string {
	if d.m.Avatar != "" {
		return d.m.Avatar
	}

	return d.UserData.Avatar()
}

func (d *MemberData) AvatarURL() string {
	if d.m.Avatar != "" {
		return d.m.AvatarURL(d.guildID)
	}

	return d.UserData.AvatarURL()
}

type CommandData struct {
	state   *statestore.Cabinet
	guildID discord.GuildID
	c       *discord.CommandInteraction
}

func NewCommandData(state *statestore.Cabinet, guildID discord.GuildID, c *discord.CommandInteraction) *CommandData {
	return &CommandData{
		state:   state,
		guildID: guildID,
		c:       c,
	}
}

func (d *CommandData) String() string {
	return d.Mention()
}

func (d *CommandData) ID() string {
	return d.c.ID.String()
}

func (d *CommandData) Name() string {
	return d.c.Name
}

func (d *CommandData) Mention() string {
	return fmt.Sprintf("</%s:%s>", d.c.Name, d.c.ID)
}

func (d *CommandData) Options() map[string]interface{} {
	res := make(map[string]interface{})
	for _, opt := range d.c.Options {
		res[opt.Name] = NewCommandOptionData(d.state, d.guildID, d.c, opt)
	}

	return res
}

func (d *CommandData) Args() map[string]interface{} {
	return d.Options()
}

func NewCommandOptionData(state *statestore.Cabinet, guildID discord.GuildID, c *discord.CommandInteraction, o discord.CommandInteractionOption) interface{} {
	switch o.Type {
	case discord.StringOptionType:
		return o.String()
	case discord.IntegerOptionType:
		v, _ := o.IntValue()
		return v
	case discord.BooleanOptionType:
		v, _ := o.BoolValue()
		return v
	case discord.UserOptionType:
		userID, _ := discord.ParseSnowflake(o.String())

		member, ok := c.Resolved.Members[discord.UserID(userID)]
		if ok {
			return NewMemberData(state, guildID, &member)
		}

		user, ok := c.Resolved.Users[discord.UserID(userID)]
		if ok {
			return NewUserData(&user)
		}

		return nil
	case discord.ChannelOptionType:
		channelID, _ := discord.ParseSnowflake(o.String())
		channel, ok := c.Resolved.Channels[discord.ChannelID(channelID)]
		if ok {
			return NewChannelData(state, channel.ID, &channel)
		}
		return NewChannelData(state, discord.ChannelID(channelID), nil)
	case discord.RoleOptionType:
		roleID, _ := discord.ParseSnowflake(o.String())
		role, ok := c.Resolved.Roles[discord.RoleID(roleID)]
		if ok {
			return NewRoleData(state, guildID, role.ID, &role)
		}
		return NewRoleData(state, guildID, discord.RoleID(roleID), nil)
	case discord.NumberOptionType:
		v, _ := o.FloatValue()
		return v
	case discord.AttachmentOptionType:
		attachmentID, _ := discord.ParseSnowflake(o.String())
		attachment, ok := c.Resolved.Attachments[discord.AttachmentID(attachmentID)]
		if ok {
			return NewAttachmentData(&attachment)
		}
		return nil
	}

	return nil
}

type GuildData struct {
	state   *statestore.Cabinet
	guildID discord.GuildID
	guild   *discord.Guild
}

func NewGuildData(state *statestore.Cabinet, guildID discord.GuildID, g *discord.Guild) *GuildData {
	return &GuildData{
		state:   state,
		guildID: guildID,
		guild:   g,
	}
}

func (d *GuildData) ensureGuild() error {
	if d.guild != nil {
		return nil
	}

	guild, err := d.state.Guild(d.guildID)
	if err != nil {
		return err
	}

	d.guild = guild
	return nil
}

func (d *GuildData) String() string {
	if err := d.ensureGuild(); err != nil {
		return d.guildID.String()
	}
	return d.guild.Name
}

func (d *GuildData) ID() string {
	return d.guildID.String()
}

func (d *GuildData) Name() (string, error) {
	if err := d.ensureGuild(); err != nil {
		return "", err
	}

	return d.guild.Name, nil
}

func (d *GuildData) Description() (string, error) {
	if err := d.ensureGuild(); err != nil {
		return "", err
	}

	return d.guild.Description, nil
}

func (d *GuildData) Icon() (string, error) {
	if err := d.ensureGuild(); err != nil {
		return "", err
	}

	return d.guild.Icon, nil
}

func (d *GuildData) IconURL() (string, error) {
	if err := d.ensureGuild(); err != nil {
		return "", err
	}

	return d.guild.IconURL(), nil
}

func (d *GuildData) Banner() (string, error) {
	if err := d.ensureGuild(); err != nil {
		return "", err
	}

	return d.guild.Banner, nil
}

func (d *GuildData) BannerURL() (string, error) {
	if err := d.ensureGuild(); err != nil {
		return "", err
	}

	return d.guild.BannerURL(), nil
}

func (d *GuildData) MemberCount() (int, error) {
	if err := d.ensureGuild(); err != nil {
		fmt.Println(err)
		return 0, err
	}

	return int(d.guild.ApproximateMembers), nil
}

func (d *GuildData) BoostCount() (int, error) {
	if err := d.ensureGuild(); err != nil {
		return 0, err
	}

	return int(d.guild.NitroBoosters), nil
}

func (d *GuildData) BoostLevel() (int, error) {
	if err := d.ensureGuild(); err != nil {
		return 0, err
	}

	return int(d.guild.NitroBoost), nil
}

type ChannelData struct {
	state     *statestore.Cabinet
	channelID discord.ChannelID
	channel   *discord.Channel
}

func NewChannelData(state *statestore.Cabinet, channelID discord.ChannelID, c *discord.Channel) *ChannelData {
	return &ChannelData{
		state:     state,
		channelID: channelID,
		channel:   c,
	}
}

func (d *ChannelData) ensureChannel() error {
	if d.channel != nil {
		return nil
	}

	channel, err := d.state.Channel(d.channelID)
	if err != nil {
		return err
	}

	d.channel = channel
	return nil
}

func (d *ChannelData) String() string {
	return d.Mention()
}

func (d *ChannelData) ID() string {
	return d.channelID.String()
}

func (d *ChannelData) Name() (string, error) {
	if err := d.ensureChannel(); err != nil {
		return "", err
	}

	return d.channel.Name, nil
}

func (d *ChannelData) Mention() string {
	return fmt.Sprintf("<#%s>", d.channelID)
}

func (d *ChannelData) Topic() (string, error) {
	if err := d.ensureChannel(); err != nil {
		return "", err
	}

	return d.channel.Topic, nil
}

type RoleData struct {
	state   *statestore.Cabinet
	guildID discord.GuildID
	roleID  discord.RoleID
	role    *discord.Role
}

func NewRoleData(state *statestore.Cabinet, guildID discord.GuildID, roleID discord.RoleID, role *discord.Role) *RoleData {
	return &RoleData{
		state:   state,
		guildID: guildID,
		roleID:  roleID,
		role:    role,
	}
}

func (d *RoleData) ensureRole() error {
	if d.role != nil {
		return nil
	}

	role, err := d.state.Role(d.guildID, d.roleID)
	if err != nil {
		return err
	}

	d.role = role
	return nil
}

func (d *RoleData) String() string {
	return d.Mention()
}

func (d *RoleData) ID() string {
	return d.roleID.String()
}

func (d *RoleData) Mention() string {
	return fmt.Sprintf("<@&%s>", d.roleID)
}

func (d *RoleData) Name() (string, error) {
	if err := d.ensureRole(); err != nil {
		return "", err
	}

	return d.role.Name, nil
}

type AttachmentData struct {
	a *discord.Attachment
}

func NewAttachmentData(a *discord.Attachment) *AttachmentData {
	return &AttachmentData{a: a}
}

func (d *AttachmentData) String() string {
	return d.URL()
}

func (d *AttachmentData) ID() string {
	return d.a.ID.String()
}

func (d *AttachmentData) URL() string {
	return d.a.URL
}

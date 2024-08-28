package placeholder

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/diamondburned/arikawa/v3/discord"
)

var ErrNotFound = errors.New("placeholder not found")

type Provider interface {
	GetPlaceholder(ctx context.Context, key string) (Provider, error)
	ResolvePlaceholder(ctx context.Context) (string, error)
}

type StringProvider struct {
	value string
}

func NewStringProvider(value string) StringProvider {
	return StringProvider{
		value: value,
	}
}

func (s StringProvider) GetPlaceholder(ctx context.Context, key string) (Provider, error) {
	return nil, ErrNotFound
}

func (s StringProvider) ResolvePlaceholder(ctx context.Context) (string, error) {
	return s.value, nil
}

type StringerProvider[T fmt.Stringer] struct {
	value T
}

func NewStringerProvider[T fmt.Stringer](value T) StringerProvider[T] {
	return StringerProvider[T]{
		value: value,
	}
}

func (s StringerProvider[T]) GetPlaceholder(ctx context.Context, key string) (Provider, error) {
	return nil, ErrNotFound
}

func (s StringerProvider[T]) ResolvePlaceholder(ctx context.Context) (string, error) {
	return s.value.String(), nil
}

type MapProvider[T Provider] struct {
	data map[string]T
}

func NewMapProvider[T Provider](data map[string]T) MapProvider[T] {
	return MapProvider[T]{
		data: data,
	}
}

func (m MapProvider[T]) Set(key string, value T) {
	m.data[key] = value
}

func (m MapProvider[T]) Delete(key string) {
	delete(m.data, key)
}

func (m MapProvider[T]) GetPlaceholder(ctx context.Context, key string) (Provider, error) {
	provider, ok := m.data[key]
	if !ok {
		return nil, ErrNotFound
	}
	return provider, nil
}

func (m MapProvider[T]) ResolvePlaceholder(ctx context.Context) (string, error) {
	return "", nil
}

type ArrayProvider[T Provider] struct {
	data []T
}

func NewArrayProvider[T Provider](data []T) ArrayProvider[T] {
	return ArrayProvider[T]{
		data: data,
	}
}

func (m ArrayProvider[T]) GetPlaceholder(ctx context.Context, key string) (Provider, error) {
	index, _ := strconv.Atoi(key)
	if index < 0 || index >= len(m.data) {
		return nil, ErrNotFound
	}

	return m.data[index], nil
}

func (m ArrayProvider[T]) ResolvePlaceholder(ctx context.Context) (string, error) {
	return "", nil
}

type InteractionProvider struct {
	interaction *discord.InteractionEvent
}

func NewInteractionProvider(interaction *discord.InteractionEvent) InteractionProvider {
	return InteractionProvider{
		interaction: interaction,
	}
}

func (p InteractionProvider) GetPlaceholder(ctx context.Context, key string) (Provider, error) {
	switch key {
	case "id":
		return NewStringProvider(p.interaction.ID.String()), nil
	case "guild":
		if p.interaction.GuildID == 0 {
			return nil, ErrNotFound
		}
		return NewGuildProvider(p.interaction.GuildID), nil
	case "channel":
		return NewChannelProvider(p.interaction.ChannelID), nil
	case "user":
		if p.interaction.Member != nil {
			return NewMemberProvider(p.interaction.Member), nil
		}
		return NewUserProvider(p.interaction.User), nil
	case "command":
		if p.interaction.Data.InteractionType() != discord.CommandInteractionType {
			return nil, ErrNotFound
		}

		return NewCommandProvider(p.interaction), nil
	}

	return nil, ErrNotFound
}

func (p InteractionProvider) ResolvePlaceholder(ctx context.Context) (string, error) {
	return p.interaction.ID.String(), nil
}

type CommandProvider struct {
	interaction *discord.InteractionEvent
	cmd         *discord.CommandInteraction
}

func NewCommandProvider(interaction *discord.InteractionEvent) CommandProvider {
	data, _ := interaction.Data.(*discord.CommandInteraction)

	return CommandProvider{
		interaction: interaction,
		cmd:         data,
	}
}

func (p CommandProvider) GetPlaceholder(ctx context.Context, key string) (Provider, error) {
	switch key {
	case "name":
		return NewStringProvider(p.cmd.Name), nil
	case "options", "args":
		res := make(map[string]CommandOptionProvider, len(p.cmd.Options))
		for _, option := range p.cmd.Options {
			res[option.Name] = NewCommandOptionProvider(p.interaction, &option)
		}

		return NewMapProvider(res), nil
	}
	return nil, ErrNotFound
}

func (p CommandProvider) ResolvePlaceholder(ctx context.Context) (string, error) {
	return p.cmd.ID.String(), nil
}

type CommandOptionProvider struct {
	interaction *discord.InteractionEvent
	option      *discord.CommandInteractionOption
}

func NewCommandOptionProvider(interaction *discord.InteractionEvent, option *discord.CommandInteractionOption) CommandOptionProvider {
	return CommandOptionProvider{
		interaction: interaction,
		option:      option,
	}
}

func (p CommandOptionProvider) GetPlaceholder(ctx context.Context, key string) (Provider, error) {
	switch key {
	case "name":
		return NewStringProvider(p.option.Name), nil
	case "value":
		return NewStringProvider(p.option.String()), nil
	}

	return nil, ErrNotFound
}

func (p CommandOptionProvider) ResolvePlaceholder(ctx context.Context) (string, error) {
	return p.option.String(), nil
}

type UserProvider struct {
	user *discord.User
}

func NewUserProvider(user *discord.User) UserProvider {
	return UserProvider{
		user: user,
	}
}

func (p UserProvider) GetPlaceholder(ctx context.Context, key string) (Provider, error) {
	switch key {
	case "id":
		return NewStringProvider(p.user.ID.String()), nil
	case "mention":
		return NewStringProvider(p.user.Mention()), nil
	case "username":
		return NewStringProvider(p.user.Username), nil
	case "discriminator":
		return NewStringProvider(p.user.Discriminator), nil
	case "display_name", "global_name", "name":
		if p.user.DisplayName != "" {
			return NewStringProvider(p.user.DisplayName), nil
		}
		return NewStringProvider(p.user.Username + "#" + p.user.Discriminator), nil
	case "avatar":
		return NewStringProvider(p.user.Avatar), nil
	case "avatar_url":
		return NewStringProvider(p.user.AvatarURL()), nil
	case "banner":
		return NewStringProvider(p.user.Banner), nil
	case "banner_url":
		return NewStringProvider(p.user.BannerURL()), nil
	}
	return nil, ErrNotFound
}

func (p UserProvider) ResolvePlaceholder(ctx context.Context) (string, error) {
	return p.user.ID.String(), nil
}

type MemberProvider struct {
	member *discord.Member
}

func NewMemberProvider(member *discord.Member) MemberProvider {
	return MemberProvider{
		member: member,
	}
}

func (p MemberProvider) GetPlaceholder(ctx context.Context, key string) (Provider, error) {
	switch key {
	case "nick":
		return NewStringProvider(p.member.Nick), nil
	case "avatar":
		return NewStringProvider(p.member.User.Avatar), nil
	case "avatar_url":
		return NewStringProvider(p.member.User.AvatarURL()), nil
	}

	return NewUserProvider(&p.member.User).GetPlaceholder(ctx, key)
}

func (p MemberProvider) ResolvePlaceholder(ctx context.Context) (string, error) {
	return p.member.User.ID.String(), nil
}

type ChannelProvider struct {
	channelID discord.ChannelID
}

func NewChannelProvider(channelID discord.ChannelID) ChannelProvider {
	return ChannelProvider{
		channelID: channelID,
	}
}

func (p ChannelProvider) GetPlaceholder(ctx context.Context, key string) (Provider, error) {
	switch key {
	case "id":
		return NewStringProvider(p.channelID.String()), nil

	}
	return nil, ErrNotFound
}

func (p ChannelProvider) ResolvePlaceholder(ctx context.Context) (string, error) {
	return p.channelID.String(), nil
}

type GuildProvider struct {
	guildID discord.GuildID
}

func NewGuildProvider(guildID discord.GuildID) GuildProvider {
	return GuildProvider{
		guildID: guildID,
	}
}

func (p GuildProvider) GetPlaceholder(ctx context.Context, key string) (Provider, error) {
	switch key {
	case "id":
		return NewStringProvider(p.guildID.String()), nil

	}
	return nil, ErrNotFound
}

func (p GuildProvider) ResolvePlaceholder(ctx context.Context) (string, error) {
	return p.guildID.String(), nil
}

package main

import (
	"encoding/json"
	"math/rand"

	kite "github.com/merlinfuchs/kite/go-sdk"
	"github.com/merlinfuchs/kite/go-sdk/discord"
	"github.com/merlinfuchs/kite/go-sdk/kv"
	"github.com/merlinfuchs/kite/go-sdk/log"
	"github.com/merlinfuchs/kite/go-types/dismodel"
	"github.com/merlinfuchs/kite/go-types/fail"
	"github.com/merlinfuchs/kite/go-types/kvmodel"
)

type Backup struct {
	Guild    dismodel.Guild     `json:"guild"`
	Channels []dismodel.Channel `json:"channels"`
	Roles    []dismodel.Role    `json:"roles"`
}

func init() {
	log.Debug("Backup Plugin loaded")
	store := kv.New("backup")

	kite.Command("backup create", func(i dismodel.Interaction, options []dismodel.ApplicationCommandOptionData) error {
		guild, err := discord.GuildGet()
		if err != nil {
			return err
		}

		channels, err := discord.ChannelList()
		if err != nil {
			return err
		}

		roles, err := discord.RoleList()
		if err != nil {
			return err
		}

		backup := Backup{
			Guild:    guild,
			Channels: channels,
			Roles:    roles,
		}

		backupRaw, err := json.Marshal(backup)
		if err != nil {
			return err
		}

		backupID := RandomString(10)
		store.Set(backupID, kvmodel.KVString(backupRaw))

		_, err = discord.InteractionResponseCreate(dismodel.InteractionResponseCreateCall{
			ID:    i.ID,
			Token: i.Token,
			Data: dismodel.InteractionResponseData{
				Content: "Backup created! ```" + backupID + "```",
			},
		})
		if err != nil {
			return err
		}

		return nil
	})

	kite.Command("backup load", func(i dismodel.Interaction, options []dismodel.ApplicationCommandOptionData) error {
		backupID := options[0].Value.(string)

		backupRaw, err := store.Get(backupID)
		if err != nil {
			if fail.IsHostErrorCode(err, fail.HostErrorTypeKVKeyNotFound) {
				_, err := discord.InteractionResponseCreate(dismodel.InteractionResponseCreateCall{
					ID:    i.ID,
					Token: i.Token,
					Data: dismodel.InteractionResponseData{
						Content: "Backup not found!",
					},
				})
				return err
			}
			return err
		}

		_, err = json.Marshal(backupRaw)
		if err != nil {
			return err
		}

		_, err = discord.InteractionResponseCreate(dismodel.InteractionResponseCreateCall{
			ID:    i.ID,
			Token: i.Token,
			Data: dismodel.InteractionResponseData{
				Content: "Backup load!",
			},
		})
		if err != nil {
			return err
		}

		return nil
	})
}

func main() {}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

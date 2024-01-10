package kite

import (
	"github.com/merlinfuchs/kite/go-sdk/internal"
	"github.com/merlinfuchs/kite/go-types/dismodel"
	"github.com/merlinfuchs/kite/go-types/event"
)

func addEventHandler(eventType event.EventType, handler event.EventHandler) {
	handlers, exists := internal.EventHandlers[eventType]
	if !exists {
		handlers = []event.EventHandler{}
	}
	handlers = append(handlers, handler)
	internal.EventHandlers[eventType] = handlers
}

func Event(eventType event.EventType, handler event.EventHandler) {
	addEventHandler(eventType, handler)
}

func Command(name string, handler event.CommandHandler) {
	addEventHandler(event.DiscordInteractionCreate, func(e event.Event) error {
		i := e.Data.(dismodel.InteractionCreateEvent)

		if i.Type != dismodel.InteractionTypeApplicationCommand {
			return nil
		}

		cmd := i.Data.(dismodel.ApplicationCommandInteractionData)

		fullCMDName := cmd.Name
		options := cmd.Options

		for _, opt := range cmd.Options {
			if opt.Type == dismodel.ApplicationCommandOptionTypeSubCommand {
				fullCMDName += " " + opt.Name
				options = opt.Options
				break
			} else if opt.Type == dismodel.ApplicationCommandOptionTypeSubCommandGroup {
				fullCMDName += " " + opt.Name
				for _, subOpt := range opt.Options {
					if subOpt.Type == dismodel.ApplicationCommandOptionTypeSubCommand {
						fullCMDName += " " + subOpt.Name
						options = subOpt.Options
						break
					}
				}
				break
			}
		}

		if fullCMDName == name {
			return handler(i, options)
		}

		return nil
	})
}

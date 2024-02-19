package kite

import (
	"slices"

	"github.com/merlinfuchs/kite/kite-sdk-go/command"
	"github.com/merlinfuchs/kite/kite-sdk-go/internal/sys"

	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-sdk-go/event"
)

func addEventHandler(eventType event.EventType, handler event.EventHandler) {
	if !slices.Contains(sys.Manifest.Events, eventType) {
		sys.Manifest.Events = append(sys.Manifest.Events, eventType)
	}

	handlers, exists := sys.EventHandlers[eventType]
	if !exists {
		handlers = []event.EventHandler{}
	}
	handlers = append(handlers, handler)
	sys.EventHandlers[eventType] = handlers
}

func Event(eventType event.EventType, handler event.EventHandler) {
	addEventHandler(eventType, handler)
}

func Command(name string) *command.Command {
	cmd := command.New(name)
	sys.Commands = append(sys.Commands, cmd)

	// TODO: consolidate into a single handler
	addEventHandler(event.DiscordInteractionCreate, func(e event.Event) error {
		i := e.Data.(distype.InteractionCreateEvent)

		if i.Type != distype.InteractionTypeApplicationCommand {
			return nil
		}

		data := i.Data.(distype.ApplicationCommandData)

		fullCMDName := cmd.Name
		options := data.Options

		for _, opt := range data.Options {
			if opt.Type == distype.ApplicationCommandOptionTypeSubCommand {
				fullCMDName += " " + opt.Name
				options = opt.Options
				break
			} else if opt.Type == distype.ApplicationCommandOptionTypeSubCommandGroup {
				fullCMDName += " " + opt.Name
				for _, subOpt := range opt.Options {
					if subOpt.Type == distype.ApplicationCommandOptionTypeSubCommand {
						fullCMDName += " " + subOpt.Name
						options = subOpt.Options
						break
					}
				}
				break
			}
		}

		if fullCMDName == cmd.FullName() {
			return cmd.Handler(i, options)
		}

		return nil
	})

	return cmd
}

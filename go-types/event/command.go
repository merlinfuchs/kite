package event

import "github.com/merlinfuchs/kite/go-types/dismodel"

type CommandHandler func(i dismodel.Interaction, options []dismodel.ApplicationCommandOptionData) error

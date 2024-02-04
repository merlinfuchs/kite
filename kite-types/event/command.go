package event

import "github.com/merlinfuchs/kite/kite-types/dismodel"

type CommandHandler func(i dismodel.Interaction, options []dismodel.ApplicationCommandOptionData) error

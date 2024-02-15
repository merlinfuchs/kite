package event

import "github.com/merlinfuchs/dismod/distype"

type CommandHandler func(i distype.Interaction, options []distype.ApplicationCommandOption) error

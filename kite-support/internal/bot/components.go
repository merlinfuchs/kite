package bot

import (
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

const (
	customIDMenuAsk        = "kite_support:menu_ask"
	customIDFollowupAsk    = "kite_support:followup_ask"
	customIDAskModal       = "kite_support:ask_modal"
	customIDFeedbackOpen   = "kite_support:feedback_open"
	customIDFeedbackModal  = "kite_support:feedback_modal"
	customIDHelpOpen       = "kite_support:help_open"
	customIDHelpModal      = "kite_support:help_modal"
	customIDQuestion       = "question"
	customIDFeedbackDetail = "details"
	customIDHelpDetail     = "details"
)

func menuEmbed() discord.Embed {
	return discord.Embed{
		Title:       "🪁 Kite Support",
		Description: "Have a question about Kite? Tap the button below and ask away — your answer is only visible to you.\n\nThe bot reads the [Kite documentation](<https://docs.kite.onl>) to find answers, so it works best for questions about features, setup, and how things work.",
		Color:       0x5865F2,
	}
}

func askModalResponse() api.InteractionResponse {
	return api.InteractionResponse{
		Type: api.ModalResponse,
		Data: &api.InteractionResponseData{
			CustomID: option.NewNullableString(customIDAskModal),
			Title:    option.NewNullableString("Ask Kite Support"),
			Components: &discord.TopLevelComponents{
				&discord.ActionRowComponent{
					&discord.TextInputComponent{
						CustomID:    customIDQuestion,
						Style:       discord.TextInputParagraphStyle,
						Label:       "Your question",
						Required:    true,
						Placeholder: "How do I create a slash command?",
					},
				},
			},
		},
	}
}

func menuComponents() *discord.TopLevelComponents {
	return &discord.TopLevelComponents{
		&discord.ActionRowComponent{
			&discord.ButtonComponent{
				Style:    discord.PrimaryButtonStyle(),
				Label:    "Ask a question",
				CustomID: customIDMenuAsk,
				Emoji:    &discord.ComponentEmoji{Name: "💬"},
			},
		},
	}
}

func followupComponents(showFeedback, showHelp bool) *discord.TopLevelComponents {
	row := discord.ActionRowComponent{
		&discord.ButtonComponent{
			Style:    discord.SecondaryButtonStyle(),
			Label:    "Ask a follow-up",
			CustomID: customIDFollowupAsk,
		},
	}
	if showFeedback {
		row = append(row, &discord.ButtonComponent{
			Style:    discord.SecondaryButtonStyle(),
			Label:    "Send as feedback",
			CustomID: customIDFeedbackOpen,
			Emoji:    &discord.ComponentEmoji{Name: "📨"},
		})
	}
	if showHelp {
		row = append(row, &discord.ButtonComponent{
			Style:    discord.SecondaryButtonStyle(),
			Label:    "Ask a human",
			CustomID: customIDHelpOpen,
			Emoji:    &discord.ComponentEmoji{Name: "🙋"},
		})
	}
	return &discord.TopLevelComponents{&row}
}

func helpModalResponse() api.InteractionResponse {
	return api.InteractionResponse{
		Type: api.ModalResponse,
		Data: &api.InteractionResponseData{
			CustomID: option.NewNullableString(customIDHelpModal),
			Title:    option.NewNullableString("Ask a human"),
			Components: &discord.TopLevelComponents{
				&discord.ActionRowComponent{
					&discord.TextInputComponent{
						CustomID:    customIDHelpDetail,
						Style:       discord.TextInputParagraphStyle,
						Label:       "What do you need help with?",
						Required:    true,
						Placeholder: "Describe what you're trying to do or what's stuck.",
					},
				},
			},
		},
	}
}

func feedbackModalResponse() api.InteractionResponse {
	return api.InteractionResponse{
		Type: api.ModalResponse,
		Data: &api.InteractionResponseData{
			CustomID: option.NewNullableString(customIDFeedbackModal),
			Title:    option.NewNullableString("Send as feedback"),
			Components: &discord.TopLevelComponents{
				&discord.ActionRowComponent{
					&discord.TextInputComponent{
						CustomID:    customIDFeedbackDetail,
						Style:       discord.TextInputParagraphStyle,
						Label:       "Anything to add?",
						Required:    false,
						Placeholder: "Optional — extra context, what you expected, etc.",
					},
				},
			},
		},
	}
}

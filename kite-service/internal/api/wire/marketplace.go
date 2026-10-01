package wire

import (
	"regexp"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/message"
	"gopkg.in/guregu/null.v4"
)

const MarketplaceMaxListingItems = 25

var discordSnowflakeRegex = regexp.MustCompile(`^[0-9]{15,21}$`)

type MarketplaceListing struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// Kind is "command", "event_listener", "message" or "module".
	Kind               string                   `json:"kind"`
	Status             string                   `json:"status"`
	Author             *MarketplaceUser         `json:"author"`
	CommandCount       int                      `json:"command_count"`
	EventListenerCount int                      `json:"event_listener_count"`
	MessageCount       int                      `json:"message_count"`
	BlockTypes         []string                 `json:"block_types"`
	ImportCount        int                      `json:"import_count"`
	Items              []MarketplaceListingItem `json:"items"`
	ReviewNote         null.String              `json:"review_note"`
	ReviewedAt         null.Time                `json:"reviewed_at"`
	CreatedAt          time.Time                `json:"created_at"`
	UpdatedAt          time.Time                `json:"updated_at"`
}

type MarketplaceListingItem struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source,omitempty"`
	// FlowSource, MessageData and MessageFlowSources are left out when
	// listings are listed.
	FlowSource *flow.FlowData `json:"flow_source,omitempty"`
	// SourceID is the message template's ID in the author's app, blocks in
	// the same listing reference the template by it.
	SourceID           string                   `json:"source_id,omitempty"`
	MessageData        *message.MessageData     `json:"message_data,omitempty"`
	MessageFlowSources map[string]flow.FlowData `json:"message_flow_sources,omitempty"`
}

// MarketplaceUser is the public part of a user, without their email.
type MarketplaceUser struct {
	ID              string      `json:"id"`
	DisplayName     string      `json:"display_name"`
	DiscordID       string      `json:"discord_id"`
	DiscordUsername string      `json:"discord_username"`
	DiscordAvatar   null.String `json:"discord_avatar"`
}

type MarketplaceListingListResponse = []*MarketplaceListing

type MarketplaceListingGetResponse = MarketplaceListing

type MarketplaceListingItemRequest struct {
	Type   string `json:"type"`
	Source string `json:"source"`
	// FlowSource is required for commands and event listeners.
	FlowSource flow.FlowData `json:"flow_source"`

	// The fields below are only used for message templates.
	Name               string                   `json:"name"`
	Description        null.String              `json:"description"`
	SourceID           string                   `json:"source_id"`
	MessageData        *message.MessageData     `json:"message_data"`
	MessageFlowSources map[string]flow.FlowData `json:"message_flow_sources"`
}

func (req MarketplaceListingItemRequest) Validate() error {
	isMessage := req.Type == string(model.MarketplaceListingItemTypeMessage)

	return validation.ValidateStruct(&req,
		validation.Field(&req.Type, validation.Required, validation.In(
			string(model.MarketplaceListingItemTypeCommand),
			string(model.MarketplaceListingItemTypeEventListener),
			string(model.MarketplaceListingItemTypeMessage),
		)),
		validation.Field(&req.Source, validation.When(
			req.Type == string(model.MarketplaceListingItemTypeEventListener),
			validation.Required,
			validation.In(string(model.EventSourceDiscord), string(model.EventSourceSchedule)),
		)),
		validation.Field(&req.FlowSource, validation.When(!isMessage, validation.Required)),
		validation.Field(&req.Name, validation.When(isMessage, validation.Required, validation.Length(1, 100))),
		validation.Field(&req.Description, validation.Length(0, 255)),
		validation.Field(&req.SourceID, validation.Length(0, 100)),
		validation.Field(&req.MessageData, validation.When(isMessage, validation.NotNil)),
	)
}

type MarketplaceListingCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	// AppID is the app the items were taken from, it's only kept for reference.
	AppID null.String                     `json:"app_id"`
	Items []MarketplaceListingItemRequest `json:"items"`
}

func (req MarketplaceListingCreateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Name, validation.Required, validation.Length(3, 64)),
		validation.Field(&req.Description, validation.Required, validation.Length(10, 2000)),
		validation.Field(&req.Items, validation.Required, validation.Length(1, MarketplaceMaxListingItems)),
	)
}

type MarketplaceListingCreateResponse = MarketplaceListing

type MarketplaceListingUpdateRequest = MarketplaceListingCreateRequest

type MarketplaceListingUpdateResponse = MarketplaceListing

type MarketplaceListingDeleteResponse = Empty

type MarketplaceListingImportResponse = MarketplaceListing

type MarketplaceListingReviewRequest struct {
	// Status is "approved", "rejected" or "removed".
	Status string `json:"status"`
	// Note is shown to the author.
	Note string `json:"note"`
}

func (req MarketplaceListingReviewRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Status, validation.Required, validation.In(
			string(model.MarketplaceListingStatusApproved),
			string(model.MarketplaceListingStatusRejected),
			string(model.MarketplaceListingStatusRemoved),
		)),
		validation.Field(&req.Note,
			validation.When(req.Status != string(model.MarketplaceListingStatusApproved), validation.Required),
			validation.Length(0, 1000),
		),
	)
}

type MarketplaceListingReviewResponse = MarketplaceListing

type MarketplaceReport struct {
	ID        string              `json:"id"`
	ListingID string              `json:"listing_id"`
	Listing   *MarketplaceListing `json:"listing"`
	Reporter  *MarketplaceUser    `json:"reporter"`
	Reason    string              `json:"reason"`
	CreatedAt time.Time           `json:"created_at"`
}

type MarketplaceReportCreateRequest struct {
	Reason string `json:"reason"`
}

func (req MarketplaceReportCreateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Reason, validation.Required, validation.Length(5, 1000)),
	)
}

type MarketplaceReportCreateResponse = Empty

type MarketplaceReportListResponse = []*MarketplaceReport

type MarketplaceReportResolveResponse = Empty

type MarketplaceModerator struct {
	DiscordUserID string `json:"discord_user_id"`
	// IsAdmin moderators come from the service config and can't be removed.
	IsAdmin bool `json:"is_admin"`
	// User is null if the moderator has never logged in to Kite.
	User      *MarketplaceUser `json:"user"`
	CreatedAt null.Time        `json:"created_at"`
}

type MarketplaceModeratorListResponse = []*MarketplaceModerator

type MarketplaceModeratorCreateRequest struct {
	DiscordUserID string `json:"discord_user_id"`
}

func (req MarketplaceModeratorCreateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.DiscordUserID, validation.Required, validation.Match(discordSnowflakeRegex).
			Error("must be a Discord user ID")),
	)
}

type MarketplaceModeratorCreateResponse = MarketplaceModerator

type MarketplaceModeratorDeleteResponse = Empty

type MarketplaceMe struct {
	IsModerator bool `json:"is_moderator"`
	IsAdmin     bool `json:"is_admin"`
	MaxListings int  `json:"max_listings"`
}

type MarketplaceMeGetResponse = MarketplaceMe

func MarketplaceUserToWire(user *model.User) *MarketplaceUser {
	if user == nil {
		return nil
	}

	return &MarketplaceUser{
		ID:              user.ID,
		DisplayName:     user.DisplayName,
		DiscordID:       user.DiscordID,
		DiscordUsername: user.DiscordUsername,
		DiscordAvatar:   user.DiscordAvatar,
	}
}

// MarketplaceListingToWire leaves out the flows of the items unless withFlows is set.
func MarketplaceListingToWire(listing *model.MarketplaceListing, author *model.User, withFlows bool) *MarketplaceListing {
	if listing == nil {
		return nil
	}

	kind := "module"
	if !listing.IsModule() {
		switch {
		case listing.CommandCount == 1:
			kind = string(model.MarketplaceListingItemTypeCommand)
		case listing.MessageCount == 1:
			kind = string(model.MarketplaceListingItemTypeMessage)
		default:
			kind = string(model.MarketplaceListingItemTypeEventListener)
		}
	}

	items := make([]MarketplaceListingItem, len(listing.Items))
	for i, item := range listing.Items {
		items[i] = MarketplaceListingItem{
			Type:        string(item.Type),
			Name:        item.Name,
			Description: item.Description,
			Source:      item.Source,
		}
		if withFlows {
			if item.Type == model.MarketplaceListingItemTypeMessage {
				items[i].SourceID = item.SourceID
				items[i].MessageData = item.MessageData
				items[i].MessageFlowSources = item.MessageFlowSources
			} else {
				flowSource := item.FlowSource
				items[i].FlowSource = &flowSource
			}
		}
	}

	blockTypes := listing.BlockTypes
	if blockTypes == nil {
		blockTypes = []string{}
	}

	return &MarketplaceListing{
		ID:                 listing.ID,
		Name:               listing.Name,
		Description:        listing.Description,
		Kind:               kind,
		Status:             string(listing.Status),
		Author:             MarketplaceUserToWire(author),
		CommandCount:       listing.CommandCount,
		EventListenerCount: listing.EventListenerCount,
		MessageCount:       listing.MessageCount,
		BlockTypes:         blockTypes,
		ImportCount:        listing.ImportCount,
		Items:              items,
		ReviewNote:         listing.ReviewNote,
		ReviewedAt:         listing.ReviewedAt,
		CreatedAt:          listing.CreatedAt,
		UpdatedAt:          listing.UpdatedAt,
	}
}

package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
)

func (c *Client) UpsertGuild(ctx context.Context, guild model.Guild) (*model.Guild, error) {
	g, err := c.Q.UpserGuild(ctx, pgmodel.UpserGuildParams{
		ID:          string(guild.ID),
		Name:        guild.Name,
		Description: nullStringToText(guild.Description),
		Icon:        nullStringToText(guild.Icon),
		CreatedAt:   timeToTimestamp(guild.CreatedAt),
		UpdatedAt:   timeToTimestamp(guild.UpdatedAt),
	})
	if err != nil {
		return nil, err
	}

	res := guildToModel(g)
	return &res, nil
}

func (c *Client) GetGuilds(ctx context.Context) ([]model.Guild, error) {
	guilds, err := c.Q.GetGuilds(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]model.Guild, len(guilds))
	for i, guild := range guilds {
		result[i] = guildToModel(guild)
	}

	return result, nil
}

func (c *Client) GetGuild(ctx context.Context, id distype.Snowflake) (*model.Guild, error) {
	guild, err := c.Q.GetGuild(ctx, string(id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, store.ErrNotFound
		}

		return nil, err
	}

	res := guildToModel(guild)
	return &res, nil
}

func (c *Client) GetDistinctGuildIDs(ctx context.Context) ([]distype.Snowflake, error) {
	ids, err := c.Q.GetDistinctGuildIDs(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]distype.Snowflake, len(ids))
	for i, id := range ids {
		result[i] = distype.Snowflake(id)
	}

	return result, nil
}

func guildToModel(guild pgmodel.Guild) model.Guild {
	return model.Guild{
		ID:          distype.Snowflake(guild.ID),
		Name:        guild.Name,
		Description: textToNullString(guild.Description),
		Icon:        textToNullString(guild.Icon),
		CreatedAt:   guild.CreatedAt.Time,
		UpdatedAt:   guild.UpdatedAt.Time,
	}
}

---
sidebar_position: 2.7
---

# Top.gg Votes

Kite can connect your bot to its listing on [Top.gg](https://top.gg). Your bot can thank or reward users when they vote, lock commands to users who voted, show its votes and post its server count to the listing.

Top.gg isn't a bot you add to your server. It's a website where bots and servers are listed and users vote for them, so everything below works through your listing's settings on Top.gg.

## Connecting Your Listing

1. On Top.gg, open the dashboard of your bot's or server's listing and go to `Integrations & API`. Copy the API token. It belongs to this listing, not to your Top.gg account.
2. In Kite, open your app, go to `Integrations`, click on `Enable` next to Top.gg and paste the token.

The token is stored encrypted and is only ever sent to Top.gg. See [Integrations](../reference/integrations.md) for more.

## Rewarding Votes

Top.gg sends a request to a webhook URL whenever someone votes, which a [webhook event listener](../reference/event.md#webhook-event-listeners) can react to.

1. Create an event listener and select `Webhook` as its source. Save it and copy its webhook URL from the settings of the first block.
2. On Top.gg, open your listing's dashboard, go to `Webhooks`, paste the URL and save. Top.gg then shows a webhook secret starting with `whs_`.
3. In Kite, save the webhook secret as a [secret](../reference/secrets.md) called `TOPGG_SECRET`.
4. Add a [Verify webhook signature](../reference/blocks/controls/control_webhook_verify.md) block, select Top.gg and enter `{{secrets.TOPGG_SECRET}}` as the key.
5. After `Invalid`, add a [Respond to webhook](../reference/blocks/actions/action_webhook_response_create.md) block with status `401`.
6. After `Valid`, add a Respond to webhook block with status `200`, then what the flow should do, like giving the voter a role or posting a thank-you in a channel.

Top.gg sends JSON like this:

```json
{
  "type": "vote.create",
  "data": {
    "weight": 1,
    "created_at": "2026-10-30T12:00:00Z",
    "expires_at": "2026-10-31T00:00:00Z",
    "user": {
      "platform_id": "123456789012345678",
      "name": "someone"
    }
  }
}
```

So the placeholders for the vote are:

| Placeholder                              | Value                                                          |
| ---------------------------------------- | -------------------------------------------------------------- |
| `{{webhook.data.data.user.platform_id}}` | Discord ID of the voter, for blocks that need a user           |
| `{{webhook.data.data.user.name}}`        | Discord username of the voter                                  |
| `{{webhook.data.data.weight}}`           | `2` on weekends, when Top.gg counts votes twice, otherwise `1` |
| `{{webhook.data.data.expires_at}}`       | When the user can vote again                                   |
| `{{webhook.data.type}}`                  | `vote.create` for a vote, `webhook.test` for a test            |

The `Send Test` button on Top.gg's webhook page sends a `webhook.test` request with you as the voter. Use a Comparison Condition on `{{webhook.data.type}}` if test votes shouldn't give rewards.

A webhook run has no server or channel, so blocks that act on a server need its ID, and messages need a channel's ID. Mention the voter with `<@{{webhook.data.data.user.platform_id}}>`. Giving a role only works if the voter is in that server.

Each app's webhooks accept bursts of up to 60 requests, then 10 a minute. Top.gg tries a vote again 3 times when it gets no answer, so very busy listings can still lose some.

## Commands Only Voters Can Use

Add a [Check Top.gg vote](../reference/blocks/actions/action_topgg_vote_get.md) block at the start of the command with the user `{{user.id}}`. Then add a Comparison Condition on `{{ result('vote').voted }}`: if it's `true`, continue with the command, otherwise reply with a link to vote, like `https://top.gg/bot/YOUR_BOT_ID/vote`.

A vote lasts 12 hours. `{{ result('vote').expires_at }}` is when the user can vote again, which you can show with the Discord timestamp of it.

## Showing Votes

[Get Top.gg listing](../reference/blocks/actions/action_topgg_project_get.md) gets the listing's votes this month (`{{ result('topgg').votes }}`), of all time (`{{ result('topgg').votes_total }}`) and its review score, for example for a `/vote` command.

## Posting the Server Count

Top.gg shows how many servers a bot is in when the bot posts it. Create an event listener with the source `Schedule`, running every hour, with a [Post Top.gg server count](../reference/blocks/actions/action_topgg_metrics_update.md) block whose server count is `{{app.guild_count}}`.

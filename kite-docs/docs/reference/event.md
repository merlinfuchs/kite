---
sidebar_position: 3
---

# Event Listener

With Event Listeners you can listen for events inside the Discord servers that your bot is in. They can also run on a [schedule](#scheduled-event-listeners) or when a [webhook](#webhook-event-listeners) is called. You find each kind on its own page under **Event Listeners** in the sidebar: **Discord Events**, **Scheduled Tasks** and **Webhooks**. Right now, Kite supports the following Discord events:

- Message Create
- Message Update
- Message Delete
- Member Join
- Member Leave
- Reaction Add
- Reaction Remove
- Bot Joined Server
- Bot Left Server
- Server Boost
- Member Banned
- Member Unbanned
- Channel Create
- Channel Update
- Channel Delete
- Member Role Add
- Member Role Remove
- Member Timeout Add
- Member Timeout Remove

## Restrictions

- By default, your app is limited to 5 Discord events. Scheduled tasks and webhooks each have their own limit.
- Kite will ignore messages that are sent by a bot.
- Reaction events ignore reactions added or removed by your own app, so a flow that reacts to a message doesn't trigger itself. Reactions by other bots still fire the event.
- Reaction events provide the emoji as `{{emoji}}`, with `{{emoji.id}}`, `{{emoji.name}}` and `{{emoji.mention}}`. `{{emoji.id}}` is empty for unicode emojis.
- Reaction Remove only knows the user's ID, so only `{{user.id}}`, `{{user.mention}}` and `{{user.created_at}}` are available there. Fields like `{{user.username}}` are empty; use a Get User block if you need them.
- Member events are only available when you enable the "Server Members Intent" in the [Discord Developer Portal](https://discord.dev).
- Bot Left Server only provides `{{guild.id}}`, as the bot no longer knows anything else about the server.
- Server Boost runs when a member boosts the server. `{{user}}` is the member who boosted, `{{guild}}` is the server, and `{{message.id}}` and `{{channel.id}}` are the boost message and the channel it was posted in.
- Kite detects boosts through the message Discord posts in the server's system messages channel. Server Boost only works when "Send a message when someone boosts this server" is enabled in the server settings and your bot can see that channel. It doesn't need the "Server Members Intent".

![Example Event Flow](./img/example-event-flow.png)

## Placeholders

The placeholder picker only lists the placeholders the event fills in. For example, Member Banned has `{{user}}` and `{{guild}}`, but no channel or message.

## Server Events

These run for changes to members and channels. Unlike reaction events, they also run for changes your own app makes, so watch out for flows that trigger themselves, like creating a channel whenever a channel is created.

### Member Banned and Member Unbanned

Run when a member is banned or unbanned. `{{user}}` is the member and `{{guild}}` is the server. Discord doesn't say who banned them or why.

The user may not be in the server, so placeholders that need a member, like `{{user.joined_at}}` or `{{user.role_names}}`, are empty. Banning a member also runs Member Leave if they were in the server.

### Member Role Add and Member Role Remove

Run when a role is added to or removed from a member. `{{user}}` is the member after the change, and `{{role}}` is the role, with `{{role.id}}`, `{{role.name}}` and `{{role.mention}}`. `{{role}}` on its own is the mention, so use `{{role.name}}` if the role shouldn't be pinged.

- When several roles change at once, the flow runs once for each of them, up to 25 times.
- To only react to one role, compare `{{role.id}}` with a Comparison Condition block.
- Kite compares the member's roles with the ones it saw last. It always knows them in servers with up to 50 members. In bigger servers it only knows members that joined, chatted or changed since your app last connected, and the first change of any other member is skipped.
- They need the "Server Members Intent" in the [Discord Developer Portal](https://discord.dev).

### Member Timeout Add and Member Timeout Remove

Member Timeout Add runs when a member is timed out, or when their timeout is changed to end at another time. `{{user.timeout_until}}` is when it ends as a Unix timestamp. Put it into `<t:{{user.timeout_until}}:R>` to show it like "in 10 minutes".

Member Timeout Remove runs when a moderator removes a member's timeout. Discord doesn't send anything when a timeout runs out by itself, so that doesn't run it.

Both need the "Server Members Intent" in the [Discord Developer Portal](https://discord.dev).

### Channel Create, Channel Update and Channel Delete

Run when a channel or category is created, updated or deleted. `{{channel}}` is the channel, with `{{channel.name}}`, `{{channel.mention}}`, `{{channel.type}}` and its category in `{{channel.category_name}}`. For a deleted channel, they're what it was right before it was deleted.

Channel Update also has the channel from before the update in `{{event.before.name}}`, `{{event.before.type}}` and `{{event.before.category_name}}`, and `{{event.changes}}` lists what changed, separated by commas: `name`, `type`, `topic`, `nsfw`, `slowmode`, `category`, `permissions`, `bitrate`, `user limit` and `region`. If Kite didn't know the channel before, these are empty.

- Threads aren't included.
- Moving a channel up or down in the list doesn't run Channel Update, as Discord sends an update for every channel that moves along with it. Moving it to another category does.

## Scheduled Event Listeners

Scheduled event listeners don't wait for something to happen in Discord. They run their flow on a schedule, which you define with a [cron expression](https://crontab.guru). Schedules always use UTC.

| Schedule       | Runs                       |
| -------------- | -------------------------- |
| `*/5 * * * *`  | Every five minutes         |
| `0 * * * *`    | At the start of every hour |
| `0 18 * * 1-5` | At 18:00 UTC on weekdays   |

A scheduled run has no user, server or channel. Blocks that act on a server, like banning a member, need a target guild, and messages need a target channel. Response blocks aren't available because there's nothing to respond to. The time the run was scheduled for is available as `{{schedule.time}}` and `{{schedule.unix}}`.

If Kite was down when a run was due, it catches up on the latest missed run if it's at most 5 minutes late. Older runs are skipped. A run is also skipped if the previous run of the same listener is still going.

### Restrictions

- Scheduled tasks have their own limit, 5 by default.
- By default, a schedule can run at most once every 5 minutes. [Premium](./premium.md) apps can run schedules once a minute.
- Every run uses credits like any other flow, so frequent schedules add up quickly. The editor shows an estimate of the credits a schedule uses per month.

## Webhook Event Listeners

Webhook event listeners run their flow when another service sends a request to their webhook URL. Use them to bring things that happen outside of Discord into your server, like a push on GitHub, a new video, a payment or a status alert.

To create one, select **Webhook** as the source of a new event listener. Open the settings of the first block in the flow editor to copy its webhook URL, or use **Copy Webhook URL** in the menu of the event listener. Then enter the URL wherever the other service asks for a webhook URL.

The URL only accepts `POST` requests. Kite answers with status `202` right away and runs the flow in the background, so the sender can't see what the flow did.

You can try it from a terminal:

```shell
curl -X POST "YOUR_WEBHOOK_URL?source=test" \
  -H "Content-Type: application/json" \
  -d '{"repository": {"name": "kite"}}'
```

### Placeholders

| Placeholder                             | Value                                                     |
| --------------------------------------- | --------------------------------------------------------- |
| `{{webhook.body}}`                      | The body of the request as text                           |
| `{{webhook.data.repository.name}}`      | A field of the body parsed as JSON, `kite` in the example |
| `{{webhook.headers['x-github-event']}}` | A header of the request, header names are in lowercase    |
| `{{webhook.query.source}}`              | A query parameter of the URL, `test` in the example       |

Many services send different kinds of events to the same URL and name the kind in a header or in the body. Use a Comparison Condition block on that value to only react to the events you want.

Like a scheduled run, a webhook run has no user, server or channel. Blocks that act on a server need a target guild, and messages need a target channel. Response blocks aren't available because there's nothing to respond to.

### Keeping the URL secret

Anyone who knows the webhook URL can run the flow, and every run uses credits. Treat the URL like a password. If it got out, regenerate it in the settings of the first block. The old URL stops working right away and the new one works within a few seconds. Enter it wherever the old one was used.

The URL isn't part of the flow, so it isn't included when you export, share or duplicate the event listener. A copy gets its own URL.

### Restrictions

- Webhooks have their own limit, which is the same as for Discord events, 5 by default.
- The body of a request can be up to 64 KB, and its headers and query parameters up to 16 KB together.
- All webhook event listeners of an app share a limit of 10 requests per minute, with bursts of up to 60 at once. Requests above that are answered with status `429` and don't run the flow.
- Requests are answered with status `404` while the event listener is disabled, and with `503` while your app is offline.
- It can take a few seconds until a new or re-enabled event listener accepts requests.

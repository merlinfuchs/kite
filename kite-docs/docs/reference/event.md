---
sidebar_position: 3
---

# Event Listener

With Event Listeners you can listen for events inside the Discord servers that your bot is in. They can also run on a [schedule](#scheduled-event-listeners) or when a [webhook](#webhook-event-listeners) is called. Right now, Kite supports the following Discord events:

- Message Create
- Message Update
- Message Delete
- Member Join
- Member Leave
- Reaction Add
- Reaction Remove
- Forum Post Create
- Forum Post Update
- Forum Post Delete
- Bot Joined Server
- Bot Left Server

## Restrictions

- By default, your app is limited to 5 event listeners.
- Kite will ignore messages that are sent by a bot.
- Reaction events ignore reactions added or removed by your own app, so a flow that reacts to a message doesn't trigger itself. Reactions by other bots still fire the event.
- Reaction events provide the emoji as `{{emoji}}`, with `{{emoji.id}}`, `{{emoji.name}}` and `{{emoji.mention}}`. `{{emoji.id}}` is empty for unicode emojis.
- Reaction Remove only knows the user's ID, so only `{{user.id}}`, `{{user.mention}}` and `{{user.created_at}}` are available there. Fields like `{{user.username}}` are empty; use a Get User block if you need them.
- Member events are only available when you enable the "Server Members Intent" in the [Discord Developer Portal](https://discord.dev).
- Forum post events only fire for posts in Discord Forum channels, not ordinary text-channel threads.
- Forum Post Create and Forum Post Update provide the post title, author ID, applied tag IDs, and thread metadata. Forum Post Delete only includes the post ID and parent Forum channel ID, so the title, author, tags, counts, and thread metadata are empty or zero. The post URL and creation time can still be derived from its ID.
- Forum post events provide `{{forum_post}}` and `{{forum}}` placeholders. The regular `{{channel}}` placeholder is the post's thread channel and `{{guild}}` is the server. When the event includes the author ID, `{{user}}`/`{{member}}` identify them with only `id`, `mention` and `created_at` available. Delete events do not include the author.
- Forum Post Update runs for thread updates, including changes to its archive or lock state.
- Forum post events don't include the starter message body as `{{message}}`; use a Message Create listener if you need to process message content.
- Bot Left Server only provides `{{guild.id}}`, as the bot no longer knows anything else about the server.

### Forum post placeholders

`{{forum_post}}` contains the event's post. `{{forum}}` is its parent Forum channel. The usual `{{channel}}` placeholder refers to the post's thread channel.

| Placeholder | Value |
| --- | --- |
| `{{forum_post.id}}` | Post thread ID |
| `{{forum_post.title}}` | Post title |
| `{{forum_post.url}}` | Discord URL for the post |
| `{{forum_post.author_id}}` | ID of the user who created the post |
| `{{forum_post.parent_channel_id}}` | ID of the parent Forum channel |
| `{{forum_post.tag_ids}}` | IDs of applied tags |
| `{{forum_post.tags}}` | Applied tag names, when the parent Forum channel is cached |
| `{{forum_post.created_at}}` | Post creation time as a Unix timestamp in seconds |
| `{{forum_post.message_count}}` | Approximate number of messages in the post thread |
| `{{forum_post.member_count}}` | Approximate number of members in the post thread |
| `{{forum_post.archived}}` | Whether the post thread is archived |
| `{{forum_post.locked}}` | Whether the post thread is locked |
| `{{forum_post.auto_archive_duration}}` | Thread auto-archive duration in minutes |

![Example Event Flow](./img/example-event-flow.png)

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

- Scheduled event listeners have their own limit, 5 by default.
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

- Webhook event listeners have their own limit, which is the same as for Discord event listeners, 5 by default.
- The body of a request can be up to 64 KB, and its headers and query parameters up to 16 KB together.
- All webhook event listeners of an app share a limit of 10 requests per minute, with bursts of up to 60 at once. Requests above that are answered with status `429` and don't run the flow.
- Requests are answered with status `404` while the event listener is disabled, and with `503` while your app is offline.
- It can take a few seconds until a new or re-enabled event listener accepts requests.

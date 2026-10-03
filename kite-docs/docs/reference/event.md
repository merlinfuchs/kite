---
sidebar_position: 3
---

# Event Listener

With Event Listeners you can listen for events inside the Discord servers that your bot is in. Right now, Kite supports the following events:

- Message Create
- Message Update
- Message Delete
- Member Join
- Member Leave
- Bot Joined Server
- Bot Left Server
- Invite Create

## Restrictions

- By default, your app is limited to 5 event listeners.
- Kite will ignore messages that are sent by a bot.
- Member events are only available when you enable the "Server Members Intent" in the [Discord Developer Portal](https://discord.dev).
- Bot Joined Server provides the server as `{{guild.id}}` and `{{guild.name}}`. Bot Left Server only provides `{{guild.id}}`.
- Invite Create needs the "Manage Channels" permission, because Discord only tells bots with it about new invites. Invites created by bots are ignored.

![Example Event Flow](./img/example-event-flow.png)

## Invite Create

Invite Create runs when someone creates an invite to a server your bot is in. The user placeholders, like `{{user.mention}}`, are the user who created the invite, and `{{channel.id}}` is the channel the invite leads to. The invite itself is available as `invite`:

| Placeholder             | Value                                                                              |
| ----------------------- | ---------------------------------------------------------------------------------- |
| `{{invite.code}}`       | The invite code, like `aBcD1234`                                                   |
| `{{invite.url}}`        | The invite link, like `https://discord.gg/aBcD1234`                                |
| `{{invite.duration}}`   | How long the invite lasts, like `7 days` or `never`                                |
| `{{invite.expires}}`    | When the invite expires as a Discord timestamp, shown like "in 7 days", or `never` |
| `{{invite.max_age}}`    | How long the invite lasts in seconds, `0` if it never expires                      |
| `{{invite.expires_at}}` | When the invite expires as a Unix timestamp, `0` if it never expires               |
| `{{invite.created_at}}` | When the invite was created as a Unix timestamp                                    |
| `{{invite.max_uses}}`   | How often the invite can be used, `0` if there is no limit                         |
| `{{invite.temporary}}`  | `true` if the invite only grants temporary membership                              |

For example, a log message could look like this:

```
{{user.mention}} created {{invite.url}} for <#{{channel.id}}>
Expires: {{invite.expires}}
Uses: {{invite.max_uses == 0 ? "unlimited" : invite.max_uses}}
```

Some invites aren't created by a user, like the one of the server widget. For those the user placeholders are empty.

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

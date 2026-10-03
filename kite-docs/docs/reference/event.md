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
- Voice Channel Join, Leave or Move

## Restrictions

- By default, your app is limited to 5 event listeners.
- Kite will ignore messages that are sent by a bot.
- Member events are only available when you enable the "Server Members Intent" in the [Discord Developer Portal](https://discord.dev).
- Bot Joined Server provides the server as `{{guild.id}}` and `{{guild.name}}`. Bot Left Server only provides `{{guild.id}}`.

![Example Event Flow](./img/example-event-flow.png)

## Voice Channel Join, Leave or Move

This event listener runs when a member joins a voice channel, leaves it, or moves to another one. It doesn't run when they mute, deafen or start streaming, and it ignores bots.

| Placeholder                | Value                                                                    |
| -------------------------- | ------------------------------------------------------------------------ |
| `{{voice.action}}`         | `joined`, `left` or `moved`                                              |
| `{{voice.channel.id}}`     | The voice channel the member is in now. Empty after they left.           |
| `{{voice.old_channel.id}}` | The voice channel the member was in before. Empty when they just joined. |

`{{user}}` is the member, and `{{channel.id}}` is the voice channel the event is about: the one they're in now, or the one they left.

The action is worded so it fits into a log message, like `{{user.mention}} {{voice.action}} <#{{channel.id}}>`. To only react to one of the three, start the flow with a `Comparison Condition` on `{{voice.action}}`. To only react to one voice channel, add an `Event Filter` on the channel ID.

Kite takes the old channel from what it has seen since your app connected. For a short moment after your app starts or reconnects, a move can show up as `joined` with an empty old channel.

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

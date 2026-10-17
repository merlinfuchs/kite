---
sidebar_position: 7
---

# Plugins

Plugins add ready-made features to your app without building any flows. They work separately from your own commands, event listeners and message templates, and they don't use any credits.

You can find them under `Plugins` in the sidebar. Right now there are five plugins: [Counting](#counting), [Starboard](#starboard), [Tickets](#tickets), [Leveling](#leveling) and [Mod Logs](#mod-logs).

## Setting Up a Plugin

1. Click on `Configure` on the plugin
2. Choose which of the plugin's commands and event listeners you want to use. All of them are selected by default.
3. Click on `Save`. The plugin's commands show up in Discord a few seconds later.

After the first setup you can turn the plugin on and off with the switch on its card. Changes take a few seconds to apply.

:::tip

Turning a plugin off doesn't remove its commands from Discord, they just stop doing anything. To remove them, turn the commands off in `Configure`.

:::

## Counting

Creates counting channels where members count up together, one message at a time.

Use `/counting-toggle` in a channel to turn counting on or off there. By default only members with the `Manage Channels` permission can use it. You can have several counting channels at the same time, and each has its own count.

- Counting starts at 1. Every correct number gets a ✅ reaction.
- Only messages that are just a number count. Other messages are ignored and don't break the count.
- If someone posts the wrong number, or two numbers in a row, the bot says who ruined it and the count starts over at 1.
- Turning counting off keeps the current count.

The bot needs the `Message Content Intent`, which you can enable in the `Bot` section of the [Discord Developer Portal](https://discord.com/developers/applications). Without it the bot can't read the numbers.

## Starboard

Copies popular messages to a starboard channel once they have enough ⭐ reactions.

Use `/starboard enable` to set it up. By default only members with the `Administrator` permission can use it.

- `channel`: The channel the messages are copied to
- `threshold`: How many reactions a message needs
- `emoji`: The reaction to count instead of ⭐ (optional)

Each server has one starboard, running `/starboard enable` again replaces its settings. Use `/starboard disable` to turn it off.

When a message reaches the threshold, the bot posts its text, author and a `Go to Message` button in the starboard channel. More reactions update the count on that post. Removing reactions doesn't lower the count, and attachments aren't copied.

## Tickets

Lets members open private support tickets with a button. Each ticket is its own channel that only the member, the support role and admins can see.

Use `/tickets setup` to post the ticket panel. By default only members with the `Manage Server` permission can use it.

- `channel`: The channel to post the panel with the `Open Ticket` button in
- `category`: The category to create the ticket channels in (optional)
- `support-role`: The role that can see and answer all tickets (optional)
- `log-channel`: The channel to log opened and closed tickets in (optional)
- `message`: The text of the panel (optional)

Running `/tickets setup` again replaces the settings and posts a new panel. Use `/tickets disable` to stop new tickets from being opened, open tickets stay and can still be closed.

Inside a ticket, anyone who can see it can use:

- `/ticket close` or the `Close Ticket` button: Deletes the ticket channel after asking to confirm
- `/ticket add`: Gives another member access to the ticket
- `/ticket remove`: Takes access away from a member again. The member who opened the ticket can't be removed.

Each member can only have one open ticket at a time. Tickets are numbered per server, like `ticket-0001`.

The bot needs the `Manage Channels` and `Manage Roles` permissions to create the ticket channels and change who can see them.

## Leveling

Gives members XP for chatting. They level up as they collect XP, and you can give out roles at certain levels.

Use `/leveling enable` to start giving XP. By default only members with the `Manage Server` permission can use the `/leveling` commands.

- `channel`: The channel to announce new levels in. Without it, new levels are announced in the channel of the message.
- `message`: The level up message. `{user}` is replaced with a mention of the member and `{level}` with their new level.

Members get 15 to 25 XP for a message, at most once a minute, so spamming doesn't level anyone up. Bots don't get XP.

- `/rank`: Shows the level, XP and progress to the next level of a member, or of yourself
- `/leaderboard`: Shows the 10 members with the most XP
- `/leveling reward-add`: Gives a role to members when they reach a level. The bot needs the `Manage Roles` permission and its own role has to be above the role it gives.
- `/leveling reward-remove`: Stops giving the role for a level. Members who already got it keep it.
- `/leveling reset`: Sets the XP of a member back to zero
- `/leveling disable`: Stops giving XP. Members keep their XP, levels and roles.

Leveling doesn't need the `Message Content Intent`, it only counts messages and doesn't read them.

## Mod Logs

Logs what happens in your server to a channel: deleted and edited messages, members joining and leaving, kicks, bans, unbans and timeouts.

Use `/modlogs enable` and pick the channel to log to. By default only members with the `Manage Server` permission can use it. Use `/modlogs disable` to stop logging.

- **Deleted and edited messages** show the old content. Discord doesn't send the content of deleted messages, so the bot remembers recent messages itself. Messages from before the bot started, or older than its last few thousand messages, are logged without their content. Messages of bots aren't logged.
- **Kicks, bans, unbans and timeouts** show the moderator and reason from the server's audit log. The bot needs the `View Audit Log` permission for that, without it they're logged without moderator and reason, and kicks look like members leaving.

You can turn off what you don't want logged in `Configure`. Keep the `message_create` event on if you log deleted or edited messages, the bot uses it to remember recent messages.

The bot needs the `Message Content Intent` to log the content of messages, and the `Server Members Intent` to log joins, leaves and timeouts. You can enable both in the `Bot` section of the [Discord Developer Portal](https://discord.com/developers/applications).

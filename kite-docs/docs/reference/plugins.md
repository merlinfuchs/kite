---
sidebar_position: 7
---

# Plugins

Plugins add ready-made features to your app without building any flows. They work separately from your own commands, event listeners and message templates, and they don't use any credits.

You can find them under `Plugins` in the sidebar. Right now there are two plugins: [Counting](#counting) and [Starboard](#starboard).

## Setting Up a Plugin

1. Click on `Configure` on the plugin
2. Choose which of the plugin's commands and event listeners you want to use. All of them are selected by default.
3. Click on `Save`
4. Click on `Deploy all commands` below the plugins so the plugin's commands show up in Discord

After the first setup you can turn the plugin on and off with the switch on its card. Changes take a few seconds to apply.

:::tip

Turning a plugin off doesn't remove its commands from Discord, they just stop doing anything. To remove them, turn the commands off in `Configure` and deploy your commands again.

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

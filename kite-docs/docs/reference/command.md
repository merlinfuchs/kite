---
sidebar_position: 1
---

# Custom command

Custom slash commands are the primary way for users to interact with your bot. Members can invoke one by typing `/` into the Discord chat.

## Context menu commands

You can also create a User or Message context menu command. In Discord, members open the menu on a user or message and choose your command. In the flow, use `interaction.command.target_user` for the selected user or `interaction.command.target_message` for the selected message. The message's author is available as `interaction.command.target_message.author`.

If a button or modal resumes the flow, the original target is available at `origin.interaction.command.target_user` or `origin.interaction.command.target_message`.

## Sub-commands

Add spaces (` `) to your command names to create sub-commands, this helps organize related commands into logical groups and improves clarity for users.

## Updating commands in Discord

Kite updates your commands in Discord a few seconds after you save them, and the `Commands` page shows when it's done. Sometimes you need to reload your Discord client (`Ctrl+R`) for the changes to appear.

Disabled commands are removed from Discord, so you can keep a command you're still working on disabled until it's ready.

If Discord rejects your commands, the `Commands` page shows why and marks the command that caused it. Discord keeps showing your commands as they were until you fix it. If your commands look out of date in Discord, click on `Sync commands` in the settings of your app.

Make sure to check your app's [logs](../guides/troubleshooting.md#logs) if your command doesn't work!

![Example Flow](./img/example-flow.webp)

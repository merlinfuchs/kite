---
sidebar_position: 1
---

# Custom Command

Custom commands are the primary way for users to interact with your bot. Once you create your first command, users will be able to use it by typing `/` into the Discord chat.

## Sub-Commands

Add spaces (` `) to your command names to create sub-commands, this helps organize related commands into logical groups and improves clarity for users.

## Updating Commands in Discord

Kite updates your commands in Discord a few seconds after you save them, and the `Commands` page shows when it's done. Sometimes you need to reload your Discord client (`Ctrl+R`) for the changes to appear.

Disabled commands are removed from Discord, so you can keep a command you're still working on disabled until it's ready.

If Discord rejects your commands, the `Commands` page shows why and marks the command that caused it. Discord keeps showing your commands as they were until you fix it. If your commands look out of date in Discord, click on `Sync Commands` in the settings of your app.

Make sure to check your app's [logs](../guides/troubleshooting.md#logs) if your command doesn't work!

![Example Flow](./img/example-flow.png)

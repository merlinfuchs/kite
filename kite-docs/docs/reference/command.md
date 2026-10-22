---
sidebar_position: 1
---

# Custom Command

Custom commands are the primary way for users to interact with your bot. Once you create your first command, users will be able to use it by typing `/` into the Discord chat.

## Command Types

When creating a command you can choose one of three types:

- **Slash Command** – the default, invoked by typing `/` in chat. Has a name, a description, and optional arguments.
- **User Context Menu** – appears when right-clicking a user and opening the **Apps** menu. Has only a name (no description or arguments).
- **Message Context Menu** – appears when right-clicking a message and opening the **Apps** menu. Has only a name (no description or arguments).

User and message context menu commands are managed in the **Context Menus** section of the Studio.

### The Right-Clicked Target

Context menu commands act on whatever the user right-clicked, available in the flow as `command.target`:

- User context menu: `{{command.target.mention}}`, `{{command.target.username}}`, `{{command.target.display_name}}`, `{{command.target.id}}`, `{{command.target.avatar_url}}`.
- Message context menu: `{{command.target.id}}`, `{{command.target.content}}`.

## Sub-Commands

Add spaces (` `) to your command names to create sub-commands, this helps organize related commands into logical groups and improves clarity for users. Sub-commands only apply to slash commands.

## Updating Commands in Discord

Kite updates your commands in Discord a few seconds after you save them, and the `Commands` page shows when it's done. Sometimes you need to reload your Discord client (`Ctrl+R`) for the changes to appear.

Disabled commands are removed from Discord, so you can keep a command you're still working on disabled until it's ready.

If Discord rejects your commands, the `Commands` page shows why and marks the command that caused it. Discord keeps showing your commands as they were until you fix it. If your commands look out of date in Discord, click on `Sync Commands` in the settings of your app.

Make sure to check your app's [logs](../guides/troubleshooting.md#logs) if your command doesn't work!

![Example Flow](./img/example-flow.png)

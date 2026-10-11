---
sidebar_position: 1
---

# Custom Command

Custom commands are the primary way for users to interact with your bot. Once you create your first command, users will be able to use it by typing `/` into the Discord chat.

## Sub-Commands

Add spaces (` `) to your command names to create sub-commands, this helps organize related commands into logical groups and improves clarity for users.

## Command Deployment

New commands only show up in Discord after you deploy them. Click on `Deploy Changes` in the editor after saving, or on `Deploy all commands` on the `Commands` page. You need to deploy again whenever you change a command's name, description, arguments or permissions. Changes to the other blocks apply a few seconds after saving, without a deploy.

You can deploy at most twice a minute. Some times it's necessary to restart or reload (ctrl+r) your Discord client for the changes to appear.

Make sure to check your app's [logs](../guides/troubleshooting.md#logs) if your command doesn't work!

![Example Flow](./img/example-flow.png)

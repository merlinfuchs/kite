---
sidebar_position: 1
---

# Custom Command

Custom commands are the primary way for users to interact with your bot. Once you create your first command, users will be able to use it by typing `/` into the Discord chat.

:::warning

Discord expects your bot to respond to an interaction within **3 seconds**
If no response is sent in time, users will see an “Interaction Failed” error.
To prevent this, you can use the `Defer Response` block to acknowledge the interaction immediately and give your bot more time to process.

:::

## Sub-Commands

Add spaces (` `) to your command names to create sub-commands, this helps organize related commands into logical groups and improves clarity for users.

## Command Deployment

Whenever you create a command or update an existing one, Kite will automatically deploy the changes to Discord within 60 seconds.
Some times it's necessary to restart or reload (ctrl+r) your Discord client for the changes to appear.

Make sure to check your app's logs in the Dashboard's overview page to see if there are any errors!

![Example Flow](./img/example-flow.png)


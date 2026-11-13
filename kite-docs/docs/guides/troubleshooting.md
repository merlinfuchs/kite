---
sidebar_position: 5
---

# Troubleshooting

## Logs

Most problems show up in your app's logs. You can find them under `Logs` in the sidebar, or in the `Logs` tab on the left side of the flow editor, which shows the last 10 logs of that command or event listener.

When a block fails, Kite logs an error with the type of the block and the reason, and the rest of the flow doesn't run. You can add your own logs with the `Log Message` block, for example to check the value of a placeholder. To handle errors yourself instead, put the blocks that can fail after a `Handle Errors` block.

Logs are kept for 30 days.

## Test a flow

The `Test` tab on the left side of the flow editor runs the flow as if someone used it. It tests the flow as it is in the editor, so you don't have to save first. On a phone, it's under the `...` button.

Fill in what the flow starts with, like the arguments of a command or the text of a message, and click `Run test`. You see every block that ran in order, with its result or error. You also see what the bot would have done, like the messages it would have sent, with a warning when Discord would reject it. That happens when the bot is missing a permission, or when a role is above the bot's highest role.

Blocks that ran get a mark on the canvas: green when they worked, yellow when Discord would likely reject what they do, and red when they failed. `Clear result` removes the marks.

Click the buttons and select menus of the messages in the result, or fill in and submit a modal, to test what happens next. The test runs again with your click, and random picks stay the same, so the bot sends the same messages as before.

Nothing is sent or changed, and tests don't cost credits. Changes to stored variables only last for the test. The bot still reads your server for real, so a wrong channel or role ID shows up as an error. Blocks that call services other than Discord, like `Send API Request` and the AI blocks, don't run in a test and have no result. `Wait` blocks are skipped.

To test what comes after such a block, click `Set test result` on it in the result. Enter what it returns, like the answer of `Ask AI` or the response of a request, or make it fail to test an error handler. For `AI Check` and `AI Choice`, pick the answer. The test runs again with it, and later tests use it too until you remove it.

To test with other stored values, like a balance that's too low, click `Start with stored values` above `Run test`. Pick the stored variable and enter the value. For variables with a value per member or server, also enter the scope with the IDs filled in, like `123_456`.

Your app has to be online to run a test.

## My bot is offline

Open the dashboard of your app. If the app was stopped, a popup tells you why.

- **Your bot token stopped working**: Discord no longer accepts the token, usually because it was reset. Click on `Reset Token` under `Bot` in the Discord Developer Portal, copy the new token and click on `Update token` in the popup or under `Credentials` in the settings.
- **Your bot needs privileged intents**: Some of your event listeners need intents that are turned off for your bot. Turn them on under `Bot` in the Discord Developer Portal, then click on `Start app` in the settings.
- **Failed to connect to gateway**: Kite couldn't connect your bot to Discord. Click on `Start app` in the settings to try again.
- **Too many servers**: Your app is in more servers than your plan allows. Remove your bot from some servers in Discord, or get a [paid plan](../reference/premium.md) with more servers, then click `Start app`.

## My bot says it's paused

Your app used all of its [credits](../reference/credit-system.md) for this month. Your bot stays online and tells anyone who uses it that it's paused, but its commands and event listeners don't run. It runs again by itself when the next month begins, or a few minutes after you upgrade.

## My command doesn't show up

Kite updates your commands in Discord a few seconds after you save them. The `Commands` page shows if that worked. If Discord rejected your commands, it shows why and which command caused it. Fix that command and save it again.

Disabled commands don't show up in Discord. Check that the command is enabled in the list of commands.

Your app has to be running for its commands to be updated. If it was stopped, start it and Kite updates them right away.

If it still doesn't show up, restart your Discord client with `Ctrl+R`, and check that your app was invited to the server with `Invite app` on the dashboard.

## My command says "The application did not respond"

Every command has to respond, usually with a `Create response message` block. Discord gives up after 3 seconds, so if your flow takes longer, add a `Defer response` block at the start. Check the logs in case a block failed before the response.

## My event listener doesn't run

- Kite ignores messages sent by bots, including your own.
- Message events need the `Message Content Intent` to see the text of messages, and member events need the `Server Members Intent`. Enable them in the `Bot` section of the [Discord Developer Portal](https://discord.com/developers/applications).
- Changes to event listeners can take up to a minute to apply after saving.
- Check that the event listener is enabled in the list of event listeners.

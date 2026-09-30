---
sidebar_position: 5
---

# Troubleshooting

## Logs

Most problems show up in your app's logs. You can find them under `Logs` in the sidebar, or in the `Logs` tab on the left side of the flow editor, which shows the last 10 logs of that command or event listener.

When a block fails, Kite logs an error with the type of the block and the reason, and the rest of the flow doesn't run. You can add your own logs with the `Log Message` block, for example to check the value of a placeholder. To handle errors yourself instead, put the blocks that can fail after a `Handle Errors` block.

Logs are kept for 30 days.

## My Bot Is Offline

Open the dashboard of your app. If the app was stopped, a popup tells you why.

- **No credits remaining**: Your app used all of its [credits](../reference/credit-system.md) for this month. Kite starts it again within a few minutes once the new month has started or you have more credits.
- **Discord bot token is invalid**: The token was reset in the Discord Developer Portal. Copy the new one and save it under `Credentials` in the settings.
- **Failed to connect to gateway**: Kite couldn't connect your bot to Discord. Click on `Start App` in the settings to try again.
- **Too many servers**: Your app is in more servers than your plan allows. Remove your bot from some servers in Discord, or get [Premium](../reference/premium.md), then click on `Start App`.

## My Command Doesn't Show Up

Commands only show up in Discord after you deploy them. Click on `Deploy Changes` in the editor, or on `Deploy all commands` on the `Commands` page. You also need to deploy again after changing a command's name, description, arguments or permissions. Changes to the blocks don't need a deploy.

Your app has to be running to deploy. If it was stopped, deploying does nothing, so start it first and deploy again.

If it still doesn't show up, restart your Discord client with `Ctrl+R`, and check that your app was invited to the server with `Invite app` on the dashboard.

## My Command Says "The application did not respond"

Every command has to respond, usually with a `Create response message` block. Discord gives up after 3 seconds, so if your flow takes longer, add a `Defer response` block at the start. Check the logs in case a block failed before the response.

## My Event Listener Doesn't Run

- Kite ignores messages sent by bots, including your own.
- Message events need the `Message Content Intent` to see the text of messages, and member events need the `Server Members Intent`. Enable them in the `Bot` section of the [Discord Developer Portal](https://discord.com/developers/applications).
- Changes to event listeners can take up to a minute to apply after saving.
- Check that the event listener is enabled in the list of event listeners.

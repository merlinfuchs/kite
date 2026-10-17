---
sidebar_position: 3
---

# Managing Your App

## Dashboard

The dashboard of your app shows how many commands, event listeners and message templates it has, how many [credits](../reference/credit-system.md) it used this month, and how many errors and warnings it logged in the last 24 hours.

Click on `Invite app` to add your bot to a server with `Add app to server`, or to your own Discord account with `Add app to account`. See [User Installable Apps](./user-installable-apps.md) for the second one. Both links use the [installation settings](#settings) of your app, including the permissions it asks for, and an option is greyed out while it's turned off there.

## Settings

You can find these on the `Settings` page of your app.

- **Start and stop**: `Stop App` takes your bot offline, and nothing it does runs until you click on `Start App`. Kite also stops your app when it runs out of credits or its token stops working.
- **Appearance**: Changes the name and description of your app. This also renames your bot in Discord. You can also upload or remove your bot's profile picture and banner here. Images can be PNG, JPEG, GIF or WebP files up to 5 MB, they change right away, and Discord only allows a few changes per hour.
- **Installation**: Chooses how people can add your app in Discord. `Server install` lets server admins add your bot to their servers, and `Account install` lets people add your app to their Discord account, see [User Installable Apps](./user-installable-apps.md). At least one has to stay on. Server installs always add your bot, and you can choose which permissions it asks for. Until you choose others, it asks for Administrator. Each turned on option shows an install link you can copy and share, and the `Invite app` button uses the same links. These are the same settings as `Installation` in the [Discord Developer Portal](https://discord.com/developers/applications).
- **Custom Status**: Sets the status and activity your bot shows in Discord, like "Playing a game". Without one, your bot shows "🪁 Powered by Kite.onl". You can add up to 10 statuses and pick which one is shown. With [Premium](../reference/premium.md) your bot can also rotate through them every minute. Every activity type except `Custom` can show an optional second line under the activity name. The stream URL is only available for the `Streaming` activity type, where Discord opens it from the `Watch` button on your bot's profile, and it has to be a Twitch or YouTube link. Changes can take a few minutes to show up.
- **Credentials**: Changes your bot's token, for example after you reset it in the [Discord Developer Portal](https://discord.com/developers/applications). The token has to belong to the same Discord app. Saving a token also starts your app again.
- **Collaborators**: Lets other people manage your app. Add them by their Discord user ID. They need to have logged in to Kite once. Collaborators can change everything except the collaborators, and they can't delete the app. The free plan doesn't include collaborators, see [Premium](../reference/premium.md).
- **Danger Zone**:
  - **Duplicate App**: Copies the app to another bot. Create a new application in the [Discord Developer Portal](https://discord.com/developers/applications) and paste the token of its bot, which can't be the bot of the app you copy or a bot that's already in Kite. The copy gets the commands, event listeners, message templates, variables, plugins, secrets and custom status of the app. It doesn't get variable values, logs, usage, collaborators or integrations, and it starts on the free plan, so an app with more than the free plan allows has to be trimmed first. Webhook event listeners get new webhook URLs. Only the owner can duplicate an app.
  - **Delete App**: Deletes the app and everything in it from Kite. You have to type the name of the app to confirm. This can't be undone, and your bot stays in the Discord Developer Portal. Only the owner can delete an app.

## Servers

The `Servers` page lists every server your bot is in. Click on the `...` next to a server and on `Leave server` to make your bot leave it.

## Emojis

The `Emojis` page lists your app's custom emojis. See [Emojis](../reference/emojis.md).

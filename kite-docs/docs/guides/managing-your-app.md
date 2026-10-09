---
sidebar_position: 3
---

# Managing Your App

## Dashboard

The dashboard of your app shows how many commands, event listeners and message templates it has, how many [credits](../reference/credit-system.md) it used this month, and how many errors and warnings it logged in the last 24 hours.

Click on `Invite app` to add your bot to a server with `Add app to server`, or to your own Discord account with `Add app to account`. See [User Installable Apps](./user-installable-apps.md) for the second one.

## Settings

You can find these on the `Settings` page of your app.

- **Start and stop**: `Stop App` takes your bot offline, and nothing it does runs until you click on `Start App`. Kite also stops your app when it runs out of credits or its token stops working.
- **Appearance**: Changes the name and description of your app. This also renames your bot in Discord.
- **Custom Status**: Sets the status and activity your bot shows in Discord, like "Playing a game". Without one, your bot shows "🪁 Powered by Kite.onl". You can add up to 10 statuses and pick which one is shown. With [Premium](../reference/premium.md) your bot can also rotate through them every minute. Every activity type except `Custom` can show an optional second line under the activity name. The stream URL is only available for the `Streaming` activity type, where Discord opens it from the `Watch` button on your bot's profile, and it has to be a Twitch or YouTube link. Changes can take a few minutes to show up.
- **Credentials**: Changes your bot's token, for example after you reset it in the [Discord Developer Portal](https://discord.com/developers/applications). The token has to belong to the same Discord app. Saving a token also starts your app again.
- **Collaborators**: Lets other people manage your app. Add them by their Discord user ID. They need to have logged in to Kite once. Collaborators can change everything except the collaborators, and they can't delete the app. The free plan doesn't include collaborators, see [Premium](../reference/premium.md).
- **Delete App**: Deletes the app and everything in it from Kite. This can't be undone. Only the owner can delete an app.

## Servers

The `Servers` page lists every server your bot is in. Click on the `...` next to a server and on `Leave server` to make your bot leave it.

## Emojis

The `Emojis` page lists your app's custom emojis. See [Emojis](../reference/emojis.md).

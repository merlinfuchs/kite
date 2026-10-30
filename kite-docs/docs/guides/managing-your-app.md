---
sidebar_position: 3
---

# Managing Your App

## Dashboard

The dashboard of your app shows how many commands, event listeners and message templates it has, how many [credits](../reference/credit-system.md) it used this month, and how many errors and warnings it logged in the last 24 hours.

Click on `Invite app` to add your bot to a server with `Add app to server`, or to your own Discord account with `Add app to account`. See [User Installable Apps](./user-installable-apps.md) for the second one.

## Settings

You can find these on the `Settings` page of your app.

- **Start and stop**: `Stop App` takes your bot offline, and nothing it does runs until you click on `Start App`. Kite also stops your app when its token stops working. When it runs out of credits, its commands and events pause until it has credits again, see [Credit System](../reference/credit-system.md).
- **Credit Messages**: Your bot sends you a direct message when it has used 80% of its credits for the month and when it has used all of them. Turn this off if you don't want them.
- **Appearance**: Changes the name and description of your app. This also renames your bot in Discord.
- **Custom Status**: Sets the status and activity your bot shows in Discord, like "Playing a game". Without one, your bot shows "🪁 Powered by Kite.onl". You can add up to 10 statuses and pick which one is shown. With [Premium](../reference/premium.md) your bot can also rotate through them every minute. Every activity type except `Custom` can show an optional second line under the activity name. The stream URL is only available for the `Streaming` activity type, where Discord opens it from the `Watch` button on your bot's profile, and it has to be a Twitch or YouTube link. Changes can take a few minutes to show up.
- **Credentials**: Changes your bot's token, for example after you reset it in the [Discord Developer Portal](https://discord.com/developers/applications). The token has to belong to the same Discord app. Saving a token also starts your app again.
- **Collaborators**: Lets other people manage your app. Add them by their Discord username or user ID. If they haven't logged in to Kite yet, add them by username and they're added when they first log in, within 7 days. Collaborators can change everything except the collaborators, and they can't delete the app. They can leave the app at the bottom of the settings page. The free plan includes 2 collaborators besides you, see [Premium](../reference/premium.md) for more.
- **Delete App**: Deletes the app and everything in it from Kite. This can't be undone. Only the owner can delete an app.

## Servers

The `Servers` page lists every server your bot is in. Click on the `...` next to a server and on `Leave server` to make your bot leave it.

With [Premium](../reference/premium.md), click on `Edit server profile` to give your bot its own nickname, avatar, banner and About Me in that server. A preview shows how the profile looks with your changes. Discord doesn't share the About Me your bot has in a server, so the preview shows its own until you type one. Fields you leave empty stay as they are, and `Reset` changes all of them back to your bot's own profile. Images can be up to 5 MB here. To change the profile from a flow, use the [Edit server profile](../reference/blocks/actions/action_server_profile_edit.md) block.

## Emojis

The `Emojis` page lists your app's custom emojis. See [Emojis](../reference/emojis.md).

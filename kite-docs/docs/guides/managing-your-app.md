---
sidebar_position: 3
---

# Managing Your App

## Dashboard

The dashboard of your app shows how many commands, event listeners and message templates it has, how many [credits](../reference/credit-system.md) it used this month, and how many errors and warnings it logged in the last 24 hours.

Click on `Invite app` to add your bot to a server with `Add app to server`, or to your own Discord account with `Add app to account`. See [User Installable Apps](./user-installable-apps.md) for the second one. Both links use the [installation settings](#settings) of your app, including the permissions it asks for, and an option is greyed out while it's turned off there.

## Settings

You can find these on the `Settings` page of your app.

- **Start and stop**: `Stop App` takes your bot offline, and nothing it does runs until you click on `Start App`. Kite also stops your app when its token stops working. When it runs out of credits, its commands and events pause until it has credits again, see [Credit System](../reference/credit-system.md).
- **Credit Messages**: Your bot sends you a direct message when it has used 80% of its credits for the month and when it has used all of them. Turn this off if you don't want them.
- **Appearance**: Changes the name and description of your app, which also renames your bot in Discord, and your bot's avatar and banner. Images can be PNG, JPEG, GIF or WebP up to 5 MB and are saved as soon as you pick them. Animated GIFs and WebPs need a paid plan. A preview next to the settings shows how your bot's profile looks in Discord.
- **Custom Status**: Sets the status and activity your bot shows in Discord, like "Playing a game". Without one, your bot shows "🪁 Powered by Kite.onl". You can add up to 10 statuses and pick which one is shown. With [Premium](../reference/premium.md) your bot can also rotate through them every minute. Every activity type except `Custom` can show an optional second line under the activity name. The stream URL is only available for the `Streaming` activity type, where Discord opens it from the `Watch` button on your bot's profile, and it has to be a Twitch or YouTube link. Changes can take a few minutes to show up.
- **Installation**: Chooses how people can add your app in Discord. `Server install` lets server admins add your bot to their servers, and `Account install` lets people add your app to their Discord account, see [User Installable Apps](./user-installable-apps.md). At least one has to stay on. Server installs always add your bot, and you can choose which permissions it asks for. Until you choose others, it asks for Administrator. Each turned on option shows an install link you can copy and share, and the `Invite app` button uses the same links. These are the same settings as `Installation` in the [Discord Developer Portal](https://discord.com/developers/applications).
- **Credentials**: Changes your bot's token, for example after you reset it in the [Discord Developer Portal](https://discord.com/developers/applications). The token has to belong to the same Discord app. Saving a token also starts your app again.
- **Collaborators**: Lets other people manage your app. Add them by their Discord username or user ID. If they haven't logged in to Kite yet, add them by username and they're added when they first log in, within 7 days. Collaborators can change everything except the collaborators, and they can't delete the app. They can leave the app at the bottom of the settings page. The free plan includes 2 collaborators besides you, see [Premium](../reference/premium.md) for more.
- **Danger Zone**:
  - **Duplicate App**: Copies the app to another bot. Create a new application in the [Discord Developer Portal](https://discord.com/developers/applications) and paste the token of its bot, which can't be the bot of the app you copy or a bot that's already in Kite. The copy gets the commands, event listeners, message templates, variables, plugins, secrets and custom status of the app. It doesn't get variable values, logs, usage, collaborators or integrations, and it starts on the free plan, so an app with more than the free plan allows has to be trimmed first. Webhook event listeners get new webhook URLs. Only the owner can duplicate an app.
  - **Delete App**: Deletes the app and everything in it from Kite. You have to type the name of the app to confirm. This can't be undone, and your bot stays in the Discord Developer Portal. Only the owner can delete an app.

## Servers

The `Servers` page lists every server your bot is in. Click on the `...` next to a server and on `Leave server` to make your bot leave it.

With [Premium](../reference/premium.md), click on `Edit server profile` to give your bot its own nickname, avatar, banner and About Me in that server. A preview shows how the profile looks with your changes. Discord doesn't share the About Me your bot has in a server, so the preview shows its own until you type one. Fields you leave empty stay as they are, and `Reset` changes all of them back to your bot's own profile. Images can be up to 5 MB here. To change the profile from a flow, use the [Edit server profile](../reference/blocks/actions/action_server_profile_edit.md) block.

## Emojis

The `Emojis` page lists your app's custom emojis. See [Emojis](../reference/emojis.md).

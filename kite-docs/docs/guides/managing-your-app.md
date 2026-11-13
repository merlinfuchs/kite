---
sidebar_position: 3
---

# Managing your app

## Dashboard

The dashboard of your app shows how many commands, event listeners and message templates it has, how many [credits](../reference/credit-system.md) it used this month, and how many errors and warnings it logged in the last 24 hours.

Click on `Invite app` to add your bot to a server with `Add app to server`, or to your own Discord account with `Add app to account`. See [User-installable apps](./user-installable-apps.md) for the second one.

## Settings

You can find these on the `Settings` page of your app.

- **Start and stop**: `Stop app` takes your bot offline, and nothing it does runs until you click `Start app`. Kite also stops your app when its token stops working. When it runs out of credits, its commands and events pause until it has credits again, see [Credit system](../reference/credit-system.md).
- **Credit messages**: Your bot sends you a direct message when it has used 80% of its credits for the month and when it has used all of them. Turn this off if you don't want them.
- **Appearance**: Changes the name and description of your app, which also renames your bot in Discord, and your bot's avatar and banner. Images can be PNG, JPEG, GIF, or WebP up to 5 MB and are saved as soon as you pick them. Animated GIFs and WebPs need Kite Plus or a higher plan. A preview next to the settings shows how your bot's profile looks in Discord.
- **Custom status**: Sets the status and activity your bot shows in Discord, like "Playing a game". Without one, your bot shows "🪁 Powered by Kite.onl". You can add up to 10 statuses and pick which one is shown. With [Kite Plus](../reference/premium.md) or a higher plan, your bot can also rotate through them every minute. Every activity type except `Custom` can show an optional second line under the activity name. The stream URL is only available for the `Streaming` activity type, where Discord opens it from the `Watch` button on your bot's profile, and it has to be a Twitch or YouTube link. Changes can take a few minutes to show up.
- **Credentials**: Changes your bot's token, for example after you reset it in the [Discord Developer Portal](https://discord.com/developers/applications). The token has to belong to the same Discord app. Saving a token also starts your app again.
- **Collaborators**: Lets other people help build your app. Click `Invite collaborator`, enter their Discord username or user ID, and pick their role. They get the invite under `Invites` in their account menu, which shows a red dot while they have one, and become a collaborator once they accept it. Invites expire after 7 days. If they haven't logged in to Kite yet, invite them by username and they see the invite when they first log in. Pending invites take up a seat and can be revoked. The free plan includes 2 collaborators besides you, and [paid plans](../reference/premium.md) include more. There are three roles:
  - **Owner**: Can do everything, picks the role of everyone they invite, can change roles, and is the only one who can invite or remove admins and delete the app.
  - **Admin**: Can change everything, including the settings. Admins can invite and remove developers, but can't remove other admins or themselves from the list. They can still leave the app.
  - **Developer**: Can edit commands, event listeners, message templates, stored variables, and everything else, but not the settings of the app, like its appearance, token, status, and collaborators. Developers also can't buy plans, make the bot leave a server, or change its server profiles.

  Collaborators can leave the app at the bottom of the settings page.
- **Audit log**: Click `View audit log` in the collaborators settings to see who changed what in your app and when, like who edited a command, invited someone or changed the bot's token. Changes made with an API token, like by an AI tool, are marked. You can filter it by the kind of change and by collaborator. Entries are kept for 90 days, and only admins and the owner can see it.
- **Delete app**: Deletes the app and everything in it from Kite. This can't be undone. Only the owner can delete an app.

## Servers

The `Servers` page lists every server your bot is in. Click on the `...` next to a server and on `Leave server` to make your bot leave it.

With [Kite Premium](../reference/premium.md) or Kite Ultimate, click `Edit server profile` to give your bot its own nickname, avatar, banner, About Me, and name style in that server. The name style is a font, an effect, and colors. Servers show the font, and the bot's profile also shows the effect and colors. A preview shows how the profile looks with your changes. Discord doesn't share the About Me your bot has in a server, so the preview shows its own until you type one. Fields you leave empty stay as they are, and `Reset` changes all of them back to your bot's own profile. Images can be up to 5 MB here. To change the profile from a flow, use the [Edit server profile](../reference/blocks/actions/action_server_profile_edit.md) block.

## Emojis

The `Emojis` page lists your app's custom emojis. See [Emojis](../reference/emojis.md).

## Installing Kite

You can install the Kite dashboard like an app, so it opens in its own window and starts on your list of apps. In Chrome or Edge on a computer, click the install icon at the end of the address bar. In Chrome on Android, open the menu and tap `Install app`. In Safari on iPhone or iPad, tap the share button and then `Add to Home Screen`.

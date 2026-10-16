---
sidebar_position: 8
---

# Emojis

Your app can have its own custom emojis, which it can use in every server it's in. You can see them under `Emojis` in the sidebar, together with their IDs.

Kite can't upload emojis itself. Click on `Create Emoji` to open the emoji page of your app in the [Discord Developer Portal](https://discord.com/developers/applications) and upload them there. New emojis can take up to a minute to show up in Kite.

## Using Emojis

Buttons and select menu options in the message editor, and blocks like `Create message reaction` and `Create poll`, have an emoji picker. Your app's emojis are listed in the picker under `Bot Emojis`.

In message text there is no picker, so you type the emoji in Discord's format instead:

- `<:name:id>` for an emoji, for example `<:kite:1234567890123456789>`
- `<a:name:id>` for an animated emoji

Use the name and ID from the `Emojis` page.

## Server Emojis

The picker also lists the custom emojis of every server your app is in, with one tab per server, so you can use them in buttons, select menus, reactions and polls too. Your app has to be online for them to show up.

Your app can use these emojis in any server, not only the one they belong to, as long as it stays in that server. If it leaves the server or the emoji is deleted, Discord shows the emoji's name instead or rejects the message.

Outside the server the emoji belongs to, Discord also needs the `Use External Emojis` permission:

- For messages your app sends, like with `Create message`, your app needs the permission in the channel.
- For responses to commands and buttons, the `@everyone` role needs the permission in the channel, because Discord checks it instead of your app's roles.

To use a server emoji in message text, get its ID by typing `\:name:` in Discord, which sends the emoji in the `<:name:id>` format.

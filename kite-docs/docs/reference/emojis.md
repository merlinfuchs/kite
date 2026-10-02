---
sidebar_position: 8
---

# Emojis

Your app can have its own custom emojis, which it can use in every server it's in. You can see them under `Emojis` in the sidebar, together with their IDs.

Kite can't upload emojis itself. Click on `Create Emoji` to open the emoji page of your app in the [Discord Developer Portal](https://discord.com/developers/applications) and upload them there. New emojis can take up to a minute to show up in Kite.

## Using Emojis

Buttons and select menu options in the message editor, and blocks like `Create message reaction` and `Create poll`, have an emoji picker. Your app's emojis are listed in the picker under `Custom Emojis`.

In message text there is no picker, so you type the emoji in Discord's format instead:

- `<:name:id>` for an emoji, for example `<:kite:1234567890123456789>`
- `<a:name:id>` for an animated emoji

Use the name and ID from the `Emojis` page. The picker only lists your app's own emojis, not the emojis of a server.

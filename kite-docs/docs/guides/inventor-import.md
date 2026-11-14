---
sidebar_position: 1.3
---

# Importing from inventor.bot

If you built commands in inventor.bot, you don't have to build them again in Kite. The import reads the slash commands of an inventor.bot bot and adds them to your Kite app as commands, which you can change like anything you built yourself.

You find it in your app under **Commands**, **Import command**, and in the checklist of a new app.

## Copy your commands out of inventor.bot

inventor.bot has no export, so a small helper copies your commands out of its dashboard.

1. On the import page, drag the **Move to Kite** button to the bookmarks bar of your browser. If you don't see the bar, press Ctrl+Shift+B, or Cmd+Shift+B on a Mac.
2. Open the bot you want to move in the inventor.bot dashboard.
3. Click the **Move to Kite** bookmark. A small box shows how many commands it found.
4. Click **Copy for Kite**, go back to Kite and paste it. If copying is blocked, click **Save as file** and upload the file in Kite instead.

The helper only runs when you click it, and only reads the commands of the bot you have open. It sends nothing anywhere: you paste what it copies into Kite yourself.

It's the same **Move to Kite** bookmark that [imports from BotGhost](./botghost-import.md). It looks at which dashboard you click it on, so you only need it once.

If your browser doesn't allow bookmarks that run code on this site, click **Can't drag it? Copy the helper instead**, open the console of the inventor.bot dashboard with F12, paste it and press Enter.

## Check what was found

Kite lists everything it found, and you pick what to import. Open an entry to see its blocks before you add it.

- **Ready**: every block has a Kite block that does the same.
- **Check after importing**: it's imported, but something needs a look first. What to check is listed below the entry, and again as a note on the canvas of the flow.
- **Can't be imported**: there's nothing in Kite to turn it into yet.

If your plan has no room for all of them, the rest is added turned off.

## What is imported

Slash commands that send a message with text are imported as a command with a response message.

## What you have to rebuild

inventor.bot doesn't document how it stores a command, so only the blocks above are converted so far. Everything else is left out and listed in a note on the canvas, so you know what to rebuild.

- **Event listeners and other triggers**, which can't be imported yet.
- **Other blocks and message parts**, like embeds, buttons and variables.

## Move the bot itself

The import copies what your bot does, not the bot. To keep the same bot in your servers, create your Kite app with the token of the bot you used in inventor.bot, and turn it off there so the two don't answer at the same time.

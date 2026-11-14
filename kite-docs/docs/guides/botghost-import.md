---
sidebar_position: 1.2
---

# Importing from BotGhost

If you built your bot in BotGhost, you don't have to build it again in Kite. The import reads the custom commands and custom events of a BotGhost bot and adds them to your Kite app as commands and event listeners, which you can change like anything you built yourself.

You find it in your app under **Commands**, **Import command**, **Import your whole bot**, and in the checklist of a new app.

## Copy your commands out of BotGhost

BotGhost has no export, and its share codes only work inside BotGhost. So a small helper copies your commands out of the BotGhost dashboard.

1. On the import page, drag the **Move to Kite** button to the bookmarks bar of your browser. If you don't see the bar, press Ctrl+Shift+B, or Cmd+Shift+B on a Mac.
2. Open the bot you want to move in the BotGhost dashboard and go to its Commands page.
3. Click the **Move to Kite** bookmark. A small box shows how many commands and custom events it found.
4. Click **Copy for Kite**, go back to Kite and paste it. If copying is blocked, click **Save as file** and upload the file in Kite instead.

The helper only runs when you click it, and only reads the commands and custom events of the bot you have open. It sends nothing anywhere: you paste what it copies into Kite yourself.

It's the same **Move to Kite** bookmark that [imports from inventor.bot](./inventor-import.md). It looks at which dashboard you click it on, so you only need it once.

## Check what was found

Kite lists everything it found, and you pick what to import. Open an entry to see its blocks before you add it.

- **Ready**: every block has a Kite block that does the same.
- **Check after importing**: it's imported, but something needs a look first. What to check is listed below the entry, and again as a note on the canvas of the flow.
- **Can't be imported**: there's nothing in Kite to turn it into, like an event Kite doesn't listen for.

If your plan has no room for all of them, the rest is added turned off. Turn them on from the Commands or Event Listeners page once there is room.

## What is imported

| BotGhost                                                        | Kite                                                                |
| --------------------------------------------------------------- | ------------------------------------------------------------------- |
| Options                                                         | Command Arguments                                                   |
| Required permissions and cooldowns of up to 1 hour              | Command Permissions and Command Cooldown                            |
| Plain text, embed and targeted replies, Send or Edit a Message  | Create response message or Create channel message                   |
| Buttons and select menus with their blocks                      | Buttons and select menus of the message, with the blocks after them |
| Reactions on a reply                                            | Create message reaction                                             |
| Send a Direct Message                                           | Send direct message                                                 |
| Random reply                                                    | Generate Random Value, then a message                               |
| Add Roles, Remove Roles                                         | Add role to member, Remove role from member                         |
| Kick, Ban, Unban, Timeout, Change Nickname                      | The moderation block of the same name                               |
| Wait                                                            | Wait                                                                |
| Send an API Request                                             | Send API Request                                                    |
| Options, Comparison, Channel, User, Role, Permissions, Chance   | Comparison, Channel and User Conditions                             |
| Variables like `{user}`, `{server}` and `{option_name}`         | [Placeholders](../reference/expressions.md)                         |

Custom events are imported for members joining and leaving, messages sent, edited and deleted, reactions added and removed, the bot joining and leaving a server, boosts, and voice channels.

## What you have to rebuild

- **Custom variables**, like `{BGVAR_coins}`, and the blocks that set them. Create a [stored variable](../reference/variable.md) and use the stored variable blocks.
- **Forms, loops, music, tickets and the built-in modules**. Kite has [templates](../reference/templates.md) for tickets and more in the Marketplace.
- **Blocks that edit or delete a message sent earlier**. They're imported as new messages, or left out.
- **Which roles, channels and users can use a command**. Set that in Discord under Server Settings, Integrations.
- **Timed events and webhooks**. Recreate them as [scheduled tasks and webhooks](../reference/event.md).

Blocks that ban, kick or time out are only imported when it's clear who they act on. Check them before you rely on them.

## Move the bot itself

The import copies what your bot does, not the bot. To keep the same bot in your servers, create your Kite app with the token of the bot you used in BotGhost, and turn it off in BotGhost so the two don't answer at the same time.

---
sidebar_position: 9
---

# Templates and Sharing

## Templates

Templates add ready-made commands and event listeners to your app, which you can then change like any you built yourself. You can find them under `Templates` in the sidebar.

- **Moderation**: The commands `/ban`, `/unban`, `/kick` and `/mute`
- **Ask AI**: An `/ask` command and an event listener that answers when someone pings your bot, both using AI. You can give the AI a personality.
- **Welcomer**: An event listener that welcomes new members in a channel. It needs the ID of that channel.
- **Modmail**: An event listener that posts the direct messages your bot receives in a staff channel, and a `/reply` command for staff to answer them. It needs the ID of the staff channel.

Click on `View details`, choose the commands and event listeners you want, fill in the inputs and click on `Import`. Deploy your commands afterwards so they show up in Discord.

## Sharing Commands, Event Listeners and Message Templates

You can copy a command, event listener or message template to another app, or share it with someone else.

1. Click on the `...` next to the command, event listener or message template and on `Export Command`, `Export Event Listener` or `Export Message`
2. Click on `Generate code` and copy the share code. You can also click on `Use JSON` to copy it as JSON instead.
3. In the other app, click on `Import command`, `Import event listener` or `Import message` and paste the code or JSON

Anyone with the code can import it. Codes expire after 90 days without use.

Message templates that a command or event listener uses are exported with it, and so are templates used by the buttons and select menus of a message template. Importing creates a copy of each of them in the new app and connects the blocks to the copies. If you import into the same app it was exported from, the templates it uses already exist there and aren't copied again. An exported message template itself is always copied.

Blocks that use a stored variable of the original app lose it when importing. Open them and pick one from the new app. Secrets and message attachments aren't exported either, so create any secrets the import warns about under `Secrets` and upload the attachments again.

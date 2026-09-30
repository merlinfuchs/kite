---
sidebar_position: 9
---

# Templates and Sharing

## Templates

Templates add ready-made commands and event listeners to your app, which you can then change like any you built yourself. You can find them under `Templates` in the sidebar.

- **Moderation**: The commands `/ban`, `/unban`, `/kick` and `/mute`
- **Ask AI**: An `/ask` command and an event listener that answers when someone pings your bot, both using AI. You can give the AI a personality.
- **Welcomer**: An event listener that welcomes new members in a channel. It needs the ID of that channel.

Click on `View details`, choose the commands and event listeners you want, fill in the inputs and click on `Import`. Deploy your commands afterwards so they show up in Discord.

## Sharing Commands and Event Listeners

You can copy a command or event listener to another app, or share it with someone else.

1. Click on the `...` next to the command or event listener and on `Export Command` or `Export Event Listener`
2. Click on `Generate code` and copy the share code. You can also click on `Use JSON` to copy the whole flow as JSON instead.
3. In the other app, click on `Import command` or `Import event listener` and paste the code or JSON

Anyone with the code can import it. Codes expire after 90 days without use.

Blocks that use a message template or stored variable still point to the one in the original app after importing. Open them and pick one from the new app instead.

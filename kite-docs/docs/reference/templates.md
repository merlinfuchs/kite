---
sidebar_position: 9
---

# Templates and Sharing

## Templates

Templates add ready-made commands and event listeners to your app, which you can then change like any you built yourself. You can find them under `Templates` in the sidebar.

- **Moderation**: The commands `/ban`, `/unban`, `/kick` and `/mute`
- **Ask AI**: An `/ask` command and an event listener that answers when someone pings your bot, both using AI. You can give the AI a personality.
- **Welcomer**: Event listeners that welcome new members in a channel you pick and say goodbye when members leave.
- **Verification**: Posts a message with a `Verify` button in a channel you pick, which gives members a role you pick. Put your bot's role above that role, and show your channels only to it.
- **Suggestions**: A `/suggest` command that posts the suggestion in a channel you pick and adds 👍 and 👎 reactions so members can vote. Members can suggest something once per minute.
- **Server Info**: The commands `/avatar`, `/userinfo` and `/serverinfo`. `/avatar` and `/userinfo` show the user who uses them unless another user is picked.
- **Modmail**: An event listener that posts the direct messages your bot receives in a staff channel, and a `/reply` command for staff to answer them. You pick the staff channel.
- **ER:LC**: The commands `/status`, `/players` and `/announce` for your [ER:LC](../guides/erlc.md) private server. Enable ER:LC under `Integrations` first.

Open a template to see what's inside, including the blocks of each command and event listener and the most [credits](./credit-system.md) each of them costs when it runs. Fill in its [parameters](./parameters.md), choose what to include and click on `Add`. Commands your app already has are left out. Everything shows up in Discord a few seconds later. Your app has each template once: if you left something out, open the template again to add it, after updating to its newest version.

Templates like Verification post a message when they're added. The channel is only used for that, so changing it later doesn't move the message. To post it somewhere else, or again after it was deleted, click `Post the Verification message again` on the template's page, then delete the old one in Discord.

When a template you added gets a new version, its page shows `Update to version`. Updating replaces the commands, event listeners and messages you didn't change with the new version. The ones you changed are kept unless you tick `replace my changes` next to them, and the ones you deleted aren't added again. Messages that were already posted in Discord keep their blocks, but use the parameter values of their message template.

`Remove from app` on the page of a template deletes everything it added that your app still has, including your changes. Message templates and stored variables that your other commands or event listeners use are kept.

## Sharing Commands, Event Listeners and Message Templates

You can copy a command, event listener or message template to another app, or share it with someone else.

1. Click on the `...` next to the command, event listener or message template and on `Export Command`, `Export Event Listener` or `Export Message`
2. Click on `Generate code` and copy the share code. You can also click on `Use JSON` to copy it as JSON instead.
3. In the other app, click on `Import command`, `Import event listener` or `Import message` and paste the code or JSON

Anyone with the code can import it. Codes expire after 90 days without use.

To share several at once, click on `Export` on the `Templates` page and choose the commands, event listeners and message templates. The message templates and stored variables they use are included. The [parameters](./parameters.md) of their flows are included without their values, so whoever imports the code fills in their own, like when adding a template. Import such a code with `Import` on the `Templates` page. What it adds are copies, later changes in the original app don't reach them.

Message templates that a command or event listener uses are exported with it, and so are templates used by the buttons and select menus of a message template. Importing creates a copy of each of them in the new app and connects the blocks to the copies. If you import into the same app it was exported from, the templates it uses already exist there and aren't copied again. An exported message template itself is always copied.

Blocks that use a stored variable of the original app lose it when importing. Open them and pick one from the new app. Secrets and message attachments aren't exported either, so create any secrets the import warns about under `Secrets` and upload the attachments again.

---
sidebar_position: 9
---

# Templates and Sharing

## Templates

Templates add ready-made commands and event listeners to your app, which you can then change like any you built yourself. You can find them under `Marketplace` in the sidebar, and on [kite.onl/marketplace](https://kite.onl/marketplace).

- **Moderation**: The commands `/ban`, `/unban`, `/kick` and `/mute`
- **Ask AI**: An `/ask` command and an event listener that answers when someone pings your bot, both using AI. You can give the AI a personality.
- **Welcomer**: Event listeners that welcome new members in a channel you pick and say goodbye when members leave.
- **Verification**: Posts a message with a `Verify` button in a channel you pick, which gives members a role you pick. Put your bot's role above that role, and show your channels only to it.
- **Suggestions**: A `/suggest` command that posts the suggestion in a channel you pick and adds 👍 and 👎 reactions so members can vote. Members can suggest something once per minute.
- **Server Info**: The commands `/avatar`, `/userinfo` and `/serverinfo`. `/avatar` and `/userinfo` show the user who uses them unless another user is picked.
- **Modmail**: An event listener that posts the direct messages your bot receives in a staff channel, and a `/reply` command for staff to answer them. You pick the staff channel.
- **ER:LC**: The commands `/status`, `/players` and `/announce` for your [ER:LC](../guides/erlc.md) private server. Enable ER:LC under `Integrations` first.
- **AFK**: An `/afk` command with an optional reason. The bot adds `[AFK]` to the member's nickname, tells anyone who pings them that they're AFK and since when, and welcomes them back when they send a message.

Open a template to see what's inside, including the blocks of each command and event listener and the most [credits](./credit-system.md) each of them costs when it runs. Fill in its [parameters](./parameters.md), choose what to include and click on `Add`. Commands your app already has are left out. Everything shows up in Discord a few seconds later. Your app has each template once: if you left something out, open the template again to add it, after updating to its newest version.

Templates like Verification post a message when they're added. The channel is only used for that, so changing it later doesn't move the message. To post it somewhere else, or again after it was deleted, click `Post the Verification message again` on the template's page, then delete the old one in Discord.

When a template you added gets a new version, its page shows `Update to version` and what changed. Updating replaces the commands, event listeners and messages you didn't change with the new version. The ones you changed are kept unless you tick `replace my changes` next to them, and the ones you deleted aren't added again. Messages that were already posted in Discord keep their blocks, but use the parameter values of their message template.

`Remove from app` on the page of a template deletes everything it added that your app still has, including your changes. Message templates and stored variables that your other commands or event listeners use are kept.

## Bundles

A bundle is like a template you make yourself. Click on `Create bundle` on the `Bundles` page, give it a name and choose its commands, event listeners and message templates. The message templates and stored variables they use are added, and the [parameters](./parameters.md) of their flows become the bundle's, without their values. A bundle can have up to 50 items.

Bundles belong to you, not to the app, so the `Bundles` page of every app shows all of yours. Open one to add it to an app, like a template. The app you made it from already has its items, so its page offers `Publish update` there instead. When you change its items, click on `Publish update` in the app you made it from, and say what changed if you like. Apps that added it then show `Update to version` with what changed, and update like a template, and items you left out of the new version stay in those apps without updates. A bundle whose app was deleted keeps working, but can't get new versions.

Until you share a bundle, only you can add or update it, even in apps you have collaborators on. To share it, click on `Share` next to it and send the link. Anyone with the link can see what's inside and add it to their app, after logging in. The page also lists what its blocks can do, like banning members, deleting messages or sending requests to other websites, because a bundle runs with the bot and credits of whoever adds it. Shared bundles aren't listed anywhere and search engines are told not to index them, only the link leads to them. Kite doesn't review bundles yet, so their page says that they're made by a user and not by Kite. Secrets and the values of parameters are never part of a bundle, but its blocks can use secrets of the same name in your app, so the page lists the secrets it reads.

Bundles that others made show up under `Added from others` on the `Bundles` page of the app you added them to. If its maker stops sharing or deletes a bundle, what it added stays in your app and keeps working, but doesn't get updates anymore. You can still remove it from its page.

## Sharing Commands, Event Listeners and Message Templates

You can copy a command, event listener or message template to another app, or share it with someone else.

1. Click on the `...` next to the command, event listener or message template and on `Export Command`, `Export Event Listener` or `Export Message`
2. Click on `Generate code` and copy the share code. You can also click on `Use JSON` to copy it as JSON instead.
3. In the other app, click on `Import command`, `Import event listener` or `Import message` and paste the code or JSON

Anyone with the code can import it. Codes expire after 90 days without use.

To share several at once, click on `Export` on the `Bundles` page and choose the commands, event listeners and message templates. The message templates and stored variables they use are included. The [parameters](./parameters.md) of their flows are included without their values, so whoever imports the code fills in their own, like when adding a template. Import such a code with `Import` on the `Bundles` page. What it adds are copies, later changes in the original app don't reach them.

Message templates that a command or event listener uses are exported with it, and so are templates used by the buttons and select menus of a message template. Importing creates a copy of each of them in the new app and connects the blocks to the copies. If you import into the same app it was exported from, the templates it uses already exist there and aren't copied again. An exported message template itself is always copied.

Blocks that use a stored variable of the original app lose it when importing. Open them and pick one from the new app. Secrets and message attachments aren't exported either, so create any secrets the import warns about under `Secrets` and upload the attachments again.

---
sidebar_position: 9
---

# Templates and Sharing

## Templates

Templates add ready-made commands and event listeners to your app. You can change them afterwards like anything you built yourself.

Find them under `Marketplace` in the sidebar, or on [kite.onl/marketplace](https://kite.onl/marketplace). The marketplace has Kite's templates and [templates users made](#your-own-templates), which Kite reviews before listing them.

Kite's templates:

- **Moderation**: The commands `/ban`, `/unban`, `/kick` and `/mute`.
- **AI Automod**: Checks every message with AI, also when it's edited, and deletes spam, scams, ads and insults. Deleted messages go to a log channel you pick. Messages the AI isn't sure about stay up and go to the log channel for a moderator to decide. You can add rules of your own. Each check costs 1 credit, so on a busy server set `New Member Days` to only check members who joined recently.
- **Ask AI**: An `/ask` command and an event listener that answers when someone pings your bot, both using AI. You can give the AI a personality.
- **Welcomer**: Welcomes new members in a channel you pick and says goodbye when members leave.
- **Verification**: Posts a message with a `Verify` button in a channel you pick, which gives members a role you pick. Put your bot's role above that role, and show your channels only to it.
- **Suggestions**: A `/suggest` command that posts the suggestion in a channel you pick with 👍 and 👎 reactions to vote. Members can suggest something once per minute.
- **Starboard**: Posts messages that get enough ⭐ reactions in a starboard channel you pick, and keeps the star count on the post up to date.
- **Server Info**: The commands `/avatar`, `/userinfo` and `/serverinfo`. `/avatar` and `/userinfo` show the user who uses them unless another user is picked.
- **Modmail**: Posts the direct messages your bot receives in a staff channel you pick, and adds a `/reply` command for staff to answer them.
- **Counting**: A channel where members count up one number at a time, also as math like `5 + 5`. The count starts over at 1 when someone posts the wrong number or two numbers in a row.
- **ER:LC**: The commands `/status`, `/players` and `/announce` for your [ER:LC](../guides/erlc.md) private server. Enable ER:LC under `Integrations` first.
- **AFK**: An `/afk` command with an optional reason. The bot adds `[AFK]` to the member's nickname, tells anyone who pings them that they're AFK and since when, and welcomes them back when they send a message.
- **Giveaways**: A `/giveaway` command for members with the Manage Server permission. It posts the giveaway with an `Enter` button, pings random winners when it ends, and has a `Reroll` button. A giveaway can run up to 30 days and be limited to a role. Delete the giveaway message to cancel it.
- **Sticky Message**: Keeps a message at the bottom of a channel you pick. Once no one has written there for 3 seconds, the bot posts it again and deletes the old one.
- **Reminders**: A `/remind` command that sends a reminder to the member's DMs after the time they pick, up to 30 days. If their DMs are closed, it pings them in the channel instead. Members can have 10 reminders waiting and set one every 30 seconds.
- **Honeypot**: Posts a warning in a channel you pick and softbans anyone who sends a message there, which catches spam bots. Members with the Manage Messages permission are ignored. Caught messages and who was kicked go to a log channel you pick. You can write the warning yourself. Put your bot's role above the members it should catch.

### Adding a Template

Open a template to see what's inside. It shows the blocks of each command and event listener, and the most [credits](./credit-system.md) each of them costs when it runs.

Fill in its [parameters](./parameters.md), choose what to include and click on `Add`. Everything shows up in Discord a few seconds later. Commands your app already has are left out.

Your app has each template once. To add something you left out, open the template again, after updating to its newest version.

Some templates, like Sticky Message, have a message you write yourself. Click on `Write message` while adding the template to open the message editor, which starts with an example. You can change it later under `Messages`, and updates of the template keep your version.

Templates like Verification post a message when they're added. The channel is only used for that, so changing it later doesn't move the message. To post it somewhere else, or again after it was deleted, click on `Post the Verification message again` on the template's page. Then delete the old one in Discord.

### Updating a Template

The `Your Templates` page lists the templates of an app under `Added to this app`. When one gets a new version, it moves to the top, and the sidebar counts it next to `Your Templates`. Its page shows `Update to version` and what changed.

Updating replaces the commands, event listeners and messages you didn't change. The ones you changed are kept, unless you tick `replace my changes` next to them. The ones you deleted aren't added again.

Messages already posted in Discord keep their blocks, but use the parameter values of their message template.

### Removing a Template

`Remove from app` on the page of a template deletes everything it added that your app still has, including your changes. Message templates and stored variables that your other commands or event listeners use are kept.

## Your Own Templates

### Creating a Template

Click on `Create template` on the `Your Templates` page, give it a name and choose its commands, event listeners and message templates. A template can have up to 50 items. The message templates and stored variables they use are added too.

The [parameters](./parameters.md) of their flows become the template's, without their values. Under `Default values` you can share the value of a text parameter, like the instructions of an AI block. Apps that add the template start with that value and can change it. Don't share anything private.

Secrets are never part of a template. Its blocks can use secrets of the same name in the app that adds it, so its page lists the secrets it reads.

Templates you make belong to you, not to the app. The `Your Templates` page of every app lists them under `Made by you`. Open one to add it to an app, or to share, list or delete it.

### Publishing an Update

When you change the items of your template, open it in the app you made it from and click on `Publish update`. You can say what changed. Apps that added it then see `Update to version`.

Items you leave out of the new version stay in those apps, but don't get updates anymore. If the app you made it from is deleted, the template keeps working but can't get new versions.

### Sharing a Link

Click on `Copy link` on the page of your template and send the link. Anyone with it can see what's inside and add it to their app after logging in.

The link can't be guessed, and search engines are told not to index it. Kite doesn't review templates shared this way, so their page says they're made by a user.

The page also lists what the blocks can do, like banning members, deleting messages or sending requests to other websites. A template runs with the bot and credits of whoever adds it, so check this before adding one.

### Listing in the Marketplace

Click on `List in marketplace` on its page, choose an icon and category, and write what it does. Kite reviews the newest version before listing it. The page shows whether it's in review, or why it wasn't approved.

Once it's listed, its link leads to its page in the marketplace. New versions are reviewed too. Until a new version is approved, apps keep getting the last approved one.

Your name on the page of a listed template leads to your creator page. It shows your Discord name and avatar and all your listed templates.

`Remove from marketplace` takes it out again. Apps that added it stay on the last approved version, and listing it again needs a new review.

### Templates Made by Others

They show up under `Added to this app` on the `Your Templates` page of the app you added them to. If the maker deletes one, what it added stays in your app and keeps working, but doesn't get updates anymore. You can still remove it from its page.

If a template breaks the rules, for example because it harms servers or collects data it shouldn't, click on `Report` on its page. Kite looks at every report, and the maker doesn't see who reported it.

A template Kite takes down is gone for everyone but its maker, who sees why and can appeal on the Discord server. Apps that added it keep what it added, but should check what it does.

## Sharing Commands, Event Listeners and Message Templates

You can copy a command, event listener or message template to another app, or share it with someone else.

1. Click on the `...` next to the command, event listener or message template and on `Export Command`, `Export Event Listener` or `Export Message`
2. Click on `Generate code` and copy the share code. You can also click on `Use JSON` to copy it as JSON instead.
3. In the other app, click on `Import command`, `Import event listener` or `Import message` and paste the code or JSON

Anyone with the code can import it. Codes expire after 90 days without use.

To share several at once, click on `Export` on the `Your Templates` page and choose the commands, event listeners and message templates. Import the code with `Import` on the `Your Templates` page of the other app. What it adds are copies, so later changes in the original app don't reach them.

The message templates and stored variables they use are included. So are message templates used by the buttons and select menus of a message template. Importing copies each of them into the new app and connects the blocks to the copies. Importing into the app it was exported from reuses the message templates that already exist there, but an exported message template itself is always copied.

The [parameters](./parameters.md) of the flows are included without their values, so whoever imports the code fills in their own, like when adding a template.

Some things aren't exported:

- Blocks that use a stored variable of the original app lose it. Open them and pick one from the new app.
- Secrets aren't exported. Create the ones the import warns about under `Secrets`.
- Message attachments aren't exported. Upload them again.

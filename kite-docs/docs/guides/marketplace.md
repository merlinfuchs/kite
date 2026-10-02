---
sidebar_position: 6
---

# Marketplace

The Marketplace is where you can share commands, event listeners and message templates with other Kite users and import what they've built into your own apps. You can find it in the `Studio` section of your app's sidebar. It replaces the old Templates page, which now lives in the `Official` tab of the Marketplace.

## Listings

Everything on the Marketplace is a listing. A listing is one of these:

- **Command**: a single command.
- **Event Listener**: a single event listener, including scheduled ones.
- **Message Template**: a single message template, including the flows of its buttons and select menus.
- **Module**: a set of up to 25 commands, event listeners and message templates that work together, for example a whole moderation or welcome system.

## Importing a Listing

1. Open the `Community` tab and search for what you need, or filter by commands, event listeners or modules.
2. Click a listing to see what it contains. Use `Preview` on each item to look at its flow. Clicking a block in the preview shows its settings.
3. Turn off the items you don't want, then click `Import`.

The items are added to the app you have open, the same way as importing a share code. They count towards your plan's limits.

Message templates are imported first. Blocks that use a message template from the same listing are pointed at the imported copy automatically, so a module keeps working out of the box. Variables, and message templates that weren't part of the listing, are cleared, so you'll need to select your own in the editor. For message templates, you can preview the message and the flow of each button or select menu before importing. Deploy your commands afterwards so they show up in Discord.

:::warning
Imported flows run with your bot's permissions. Listings that use blocks which can remove things in your server (like banning members or deleting channels) or send data outside of Discord (like the HTTP request block) show a warning. Check what those blocks do before enabling them.
:::

## Publishing a Listing

1. Click `Publish` on the Marketplace page.
2. Give the listing a name and a description that explains what it does and how to set it up.
3. Select the commands, event listeners and message templates from your current app that you want to include. Selecting more than one publishes them together as a module.

If a selected block uses a message template that isn't selected, you'll be asked to include it, otherwise the block loses its template when someone imports the listing. Attachments of message templates aren't published, since they are stored in your app.

Anyone can read the flows you publish, so remove API keys, tokens and private URLs from your blocks first.

New listings are reviewed by a moderator before they become public. Until then, only you can see them in the `My Listings` tab. If a listing is rejected or removed, the moderator's note is shown there so you know what to change. You can edit your listing at any time, either keeping the current contents or replacing them with items from your app. Every edit is reviewed again before it goes public.

## Reporting a Listing

If a listing is malicious, broken or breaks the rules, open it and click `Report`. Listings that receive several reports are hidden automatically until a moderator has looked at them.

## Moderation

Marketplace moderators can review new listings, handle reports and take down listings from the `Moderation` page, which is linked from the Marketplace for moderators only. Moderators are identified by their Discord user ID, so they can be added before they have logged in to Kite.

When reviewing a listing, open the preview of every item and check what its blocks do. A note is required when rejecting or removing a listing and is shown to the author. Reviewing a listing resolves all of its open reports.

### Self-hosting

If you host Kite yourself, the Marketplace is configured in the `[marketplace]` section of `kite.toml`:

```toml
[marketplace]
# Discord user IDs that can moderate and add or remove other moderators.
admin_discord_ids = ["123456789012345678"]
# Hide new and changed listings until a moderator approves them.
require_review = true
# How many listings one user can publish, 0 is unlimited.
max_listings_per_user = 25
# How many open reports hide an approved listing, 0 disables it.
auto_hide_reports = 3
```

Discord IDs must be quoted. With environment variables, separate multiple IDs with commas, for example `KITE_MARKETPLACE__ADMIN_DISCORD_IDS="123456789012345678,876543210987654321"`.

Admins can add and remove other moderators from the `Moderators` tab of the moderation page. Listings published by moderators and admins don't need a review.

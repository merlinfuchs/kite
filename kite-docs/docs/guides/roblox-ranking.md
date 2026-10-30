---
sidebar_position: 2.6
---

# Roblox Ranking Bot

Kite can connect your Discord bot to your Roblox group. Staff can rank members from Discord, anyone can look up a member's rank, and verified members can let themselves into the group.

If you don't have a bot yet, create one first with the [getting started guide](./getting-started.md).

Kite is a third-party app and isn't affiliated with Roblox or Bloxlink. For help with your Kite bot, ask in [Kite's Discord server](https://discord.gg/rNd9jWHnXh).

## Connecting Bloxlink

Roblox doesn't know which Discord account belongs to which Roblox account. Kite finds out through [Bloxlink](https://blox.link), the bot members use to verify their Roblox account in your server.

1. Add the Bloxlink bot to your server, if it isn't there yet, and have members verify with it.
2. Get your server's API key from the [Bloxlink dashboard](https://blox.link/dashboard).
3. In Kite, open your app, go to `Integrations`, click on `Enable` next to Bloxlink and paste the key.

The key only works for your server, so Kite looks members up in the server the command runs in.

## Connecting Your Group

Changing roles needs an Open Cloud API key from Roblox. You don't give Kite your password or cookie.

1. Pick the account the key belongs to. It has to be able to change members' roles in the group, and Roblox only lets it give roles below its own highest role. An account with a staff role instead of the owner's limits what the key can do.
2. Log in with that account on [Roblox's Creator Hub](https://create.roblox.com/dashboard/credentials) and create an API key under _Credentials_.
3. Under _Access Permissions_, add _groups_ with the _group:read_ and _group:write_ permissions.
4. Under _Accepted IP Addresses_, add `0.0.0.0/0`, so Kite can use the key.
5. In Kite, go to `Integrations`, click on `Enable` next to Roblox Open Cloud and paste the key.

Both keys are stored encrypted and only sent to Bloxlink and Roblox. See [Integrations](../reference/integrations.md) for more.

## Adding the Roblox Ranking Template

Go to `Marketplace` in the sidebar, click on Roblox Ranking and enter your group's ID, the number in its link. You get three commands:

- `/setrank` changes a member's role to the one you name, like `/setrank @member Moderator`. The role can also be its rank number. Only members with the `Manage Roles` permission can use it.
- `/getrank` shows a member's role in the group, or your own.
- `/joingroup` accepts your request to join the group. Members send the request on Roblox first.

If a member hasn't verified with Bloxlink, the commands tell them to. If one fails, it replies with the error, like Roblox saying the key's account can't give that role.

## Building Your Own Commands

The template commands are normal commands, so you can open them in the flow editor and change them. These blocks talk to Bloxlink and your group:

- [Get Bloxlink User](../reference/blocks/actions/action_bloxlink_user_get.md) finds the Roblox account of a Discord member.
- [Get Roblox Group Member](../reference/blocks/actions/action_roblox_group_member_get.md) gets a user's roles in the group.
- [Change Roblox Group Role](../reference/blocks/actions/action_roblox_group_member_role_edit.md) ranks a member, or gives or takes one of their roles.
- [Answer Roblox Join Request](../reference/blocks/actions/action_roblox_group_join_request_respond.md) accepts or declines a request to join.

The group blocks take a Roblox ID or username. Pass the result of Get Bloxlink User, like `{{ result('id') }}`, to use the account a member verified with.

For example, to give members a Discord role that matches their rank, add a Get Roblox Group Member block after Get Bloxlink User, then a [Comparison Condition](../reference/blocks/controls/control_condition_compare.md) on `{{ result('id').rank }}`, with an [Add role to member](../reference/blocks/actions/action_member_role_add.md) block for each rank.

Groups can give members several roles. Change Roblox Group Role replaces all of them by default, like a rank. Set its mode to _Add role_ or _Remove role_ to change one and keep the others.

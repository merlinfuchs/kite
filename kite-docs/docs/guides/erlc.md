---
sidebar_position: 2.5
---

# ER:LC Discord Bot

Kite can connect your Discord bot to your [Emergency Response: Liberty County](https://apidocs.erlc.gg) private server. Your bot can show the server's status and players, run commands like `:m` or `:kick` from Discord, and post what happens in the game to a channel.

If you don't have a bot yet, create one first with the [getting started guide](./getting-started.md).

Kite is a third-party app and isn't affiliated with ER:LC or Police Roleplay Community. For help with your Kite bot, ask in [Kite's Discord server](https://discord.gg/rNd9jWHnXh).

## Connecting Your Server

1. In ER:LC, open the settings of your private server and copy its server key. ER:LC explains where to find it in [their docs](https://apidocs.erlc.gg/how-to-obtain-your-server-key).
2. In Kite, open your app, go to `Integrations`, click on `Enable` next to ER:LC and paste the server key.
3. Click on `Authorize Kite` and confirm it on ER:LC's page. Your bot can read the server's status without it, but ER:LC only runs commands from apps the server owner authorized.

The key is stored encrypted and is only ever sent to ER:LC. See [Integrations](../reference/integrations.md) for more.

## Importing the ER:LC Template

The quickest start is the ER:LC template. Go to `Templates` in the sidebar, click on `View details` under ER:LC and import the commands you want:

- `/status` shows the server's name, player count and join code.
- `/players` lists everyone in the server with their team.
- `/announce` shows a message to everyone in the server with `:m`. Only members with the `Manage Server` permission can use it.

The commands show up in Discord a few seconds later. If one fails, it replies with ER:LC's error, like a server key that isn't valid anymore.

## Building Your Own Commands

The template commands are normal commands, so you can open them in the flow editor and change them. Two blocks talk to ER:LC:

- [Get ER:LC server](../reference/blocks/actions/action_erlc_server_get.md) gets the server's status. Turn on the parts you need, like players, staff, the queue, join logs or kill logs. ER:LC only returns those.
- [Run ER:LC command](../reference/blocks/actions/action_erlc_command_run.md) runs a command in the server, like `:h Server restart in 5 minutes`.

The results of Get ER:LC server show up in the placeholder explorer, so you can click on a field like the player count instead of typing it. `result('id')` stands for the result of the block with that ID, which the placeholder explorer fills in for you. With players and the queue turned on, you can for example count them:

```
Police online: {{ len(filter(result('id').Players, #.Team == 'Police')) }}
Waiting to join: {{ len(result('id').Queue) }}
```

Players are written as `name:id`, like `Roblox:1`. `split(#.Player, ':')[0]` gets just the name, as the `/players` command does.

For a command that kicks a player, add a [Command Argument](../reference/blocks/options/option_command_argument.md) called `player`, then a Run ER:LC command block with the command `:kick {{arg('player')}}`. Add a [Command Permissions](../reference/blocks/options/option_command_permissions.md) block, so only your staff can use it.

ER:LC runs one command every 5 seconds per server. If another command just ran, the block waits for its turn, up to 10 seconds. Don't run commands in a loop, ER:LC's rules forbid spamming and breaking them can cost the server its API access.

## Getting Events from the Game

ER:LC can send your bot a request when something happens in the game. Right now it does that for emergency calls and for chat messages starting with `;`, which lets you make your own in-game commands.

1. Create an event listener and select `Webhook` as its source.
2. Add a [Verify webhook signature](../reference/blocks/controls/control_webhook_verify.md) block and select ER:LC. It needs no key.
3. After `Invalid`, add a [Respond to webhook](../reference/blocks/actions/action_webhook_response_create.md) block with status `401`. ER:LC only accepts a webhook URL that rejects requests it didn't sign.
4. After `Valid`, add a Respond to webhook block with status `200`, then what the flow should do. To see what ER:LC sends, start with a [Create channel message](../reference/blocks/actions/action_message_create.md) block that posts `{{webhook.body}}` to a staff channel. Enter the channel's ID, as a webhook run has no channel of its own.
5. Copy the webhook URL from the settings of the first block. In ER:LC, open the settings of your private server, search for `Event Webhook` and paste it.

ER:LC tests the URL when you save it, so save the event listener in Kite first. Once you know what the requests look like, pick out fields with `{{webhook.data...}}` and react to the ones you want with a Comparison Condition. See [Webhook Event Listeners](../reference/event.md#webhook-event-listeners) for all placeholders and limits.

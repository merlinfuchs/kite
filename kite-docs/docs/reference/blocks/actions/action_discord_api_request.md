---
sidebar_position: 35.1
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Discord API Request

<EmbedFlowNode type="action_discord_api_request" />

The `Discord API Request` block calls an endpoint of the [Discord API](https://discord.com/developers/docs/reference) as your bot. Use it for things no other block does yet, like listing the messages of a channel.

Pick the endpoint from the list and fill in its parameters. Kite adds your bot's token to the request, so never paste the token into a flow. Endpoints are named after Discord's [OpenAPI spec](https://github.com/discord/discord-api-spec), which sometimes differs from the docs: _Modify Guild_ is `update_guild` and _Get Channel Messages_ is `list_messages`.

Placeholders work in the parameters and in the string values of the JSON body. A value that is only a placeholder keeps its type, so `"limit": "{{ 5 }}"` sends the number `5`. To send it as text instead, use `{{ string(5) }}`.

The block's result is the JSON the endpoint returns, so you can read its fields directly, e.g. `{{ result('list')[0].content }}` for the first message of _List Messages_. A request that Discord rejects fails the block with Discord's error message.

In flows started by an interaction, like a command or a button click, `create_followup_message`, `update_original_interaction_response` and `update_followup_message` send a followup or edit a response of that interaction, for messages the response blocks can't build, like an ephemeral reply whose fields come from a list. Kite fills in the app's ID and the interaction's token. If nothing has responded yet, Kite defers first, and ephemerally if the body has `"flags": 64`. The token is only valid for 15 minutes after the interaction, and after a button click it's the click's, not the command's.

Some endpoints aren't available because Kite manages what they change or they don't use the bot's token: changing the bot's name and avatar, the app's commands, emojis and settings, the first response to an interaction and webhooks that use a token. File uploads aren't supported.

<NodeInfoExplorer type="action_discord_api_request" />

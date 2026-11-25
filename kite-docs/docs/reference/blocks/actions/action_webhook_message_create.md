---
sidebar_position: 9.6
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Send webhook message

<EmbedFlowNode type="action_webhook_message_create" />

The `Send webhook message` block sends a message through a Discord webhook, which can be one you made with the [Create webhook](./action_webhook_create.md) block or any other webhook, even one of another server. Set the `Webhook URL` to its URL, like `https://discord.com/api/webhooks/ID/TOKEN`. The URL has the webhook's token, so you can store it as a secret and use `{{secrets.WEBHOOK_URL}}`. Your bot doesn't need any permission for it.

The message is sent with the name and avatar of the webhook. Set `Thread ID` to send it to a thread of the webhook's channel instead of the channel.

Buttons and select menus only work on messages of webhooks that your app created, like those of the [Create webhook](./action_webhook_create.md) block. Discord doesn't allow names that contain `discord` or `clyde`. The [Edit channel message](./action_message_edit.md) block can't edit messages sent with this block, as only the webhook can.

To send with any name and avatar per message without managing a webhook yourself, use [Send message as](./action_persona_message_create.md) instead, which makes one for you.

<NodeInfoExplorer type="action_webhook_message_create" />

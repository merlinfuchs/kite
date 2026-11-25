---
sidebar_position: 26.3
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Create webhook

<EmbedFlowNode type="action_webhook_create" />

The `Create webhook` block creates a webhook in a channel, which your bot needs the `Manage Webhooks` permission there for. Leave the channel empty to use the channel the flow runs in. Threads and forums can't have webhooks of their own, create it in their parent channel.

The result is the created webhook, with its `id`, `token`, `name` and `channel_id`. Its URL is `https://discord.com/api/webhooks/{{ result('webhook').id }}/{{ result('webhook').token }}`, which you can send messages to with the [Send webhook message](./action_webhook_message_create.md) block.

The token is a secret: anyone who has it can send messages through the webhook. Don't post it in a channel, and store it in a [secret](../../secrets.md) or a [variable](../../variable.md) if you need it later.

<NodeInfoExplorer type="action_webhook_create" />

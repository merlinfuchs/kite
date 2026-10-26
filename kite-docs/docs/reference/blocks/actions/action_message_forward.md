---
sidebar_position: 12.05
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Forward message

<EmbedFlowNode type="action_message_forward" />

The `Forward message` block is used to forward a message to another channel, like the Forward option in Discord. The forward shows the original message with its content, embeds and attachments.

Set `Target Message` to the ID of the message to forward, for example `{{message.id}}` in a message event listener, and `Forward To` to the channel to send the forward to. Leave `Message Channel` empty to forward a message from the channel the flow runs in.

The bot needs permission to read the original message and to send messages in the channel it forwards to. The forwarded message is the block's result, so `{{result('forward_xxx').id}}` is its ID.

<NodeInfoExplorer type="action_message_forward" />

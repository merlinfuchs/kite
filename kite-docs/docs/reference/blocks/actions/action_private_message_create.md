---
sidebar_position: 13
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Send direct message

<EmbedFlowNode type="action_private_message_create" />

The `Send direct message` block is used to send a message to a specific user.

You can either configure the message in the block directly or use a message template instead. In both cases you can add embeds and interactive components to the message. The only case where it's better to use a message template is when you want to use the same response in multiple places.

## Users with DMs turned off

Users can turn off direct messages from server members or block your bot, and Discord also refuses DMs to users who share no server with it. When the message can't be delivered for one of these reasons, the block doesn't stop the flow: the blocks connected after it still run, and a warning is added to your app's logs.

The block's result is empty in that case. To do something different when the DM didn't go through, for example to reply in the channel instead, add a `Comparison Condition` after the block with `{{result('BLOCK_ID') == nil}}` as the base value and a branch that checks if it equals `true`.

Other errors, like an invalid user ID, still stop the flow as usual.

<NodeInfoExplorer type="action_private_message_create" />

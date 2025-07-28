---
sidebar_position: 13
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Send direct message

<EmbedFlowNode type="action_private_message_create" />

The `Send direct message` block is used to send a message to a specific user.

You can either configure the message in the block directly or use a message template instead. In both cases you can add embeds and interactive components to the message. The only case where it's better to use a message template is when you want to use the same response in multiple places.

<NodeInfoExplorer type="action_private_message_create" />

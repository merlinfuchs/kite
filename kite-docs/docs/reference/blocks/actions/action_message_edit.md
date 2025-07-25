---
sidebar_position: 10
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";

# Edit channel message

The `Edit channel message` block is used to edit a message in a specific channel.

You can either configure the message in the block directly or use a message template instead. In both cases you can add embeds and interactive components to the message. The only case where it's better to use a message template is when you want to use the same response in multiple places.

If the message contains interactive components, the flow will be suspended until the user interacts with the message. See [Sub-Flows](/reference/sub-flows) for more information on how interactive components work.

<EmbedFlowNode type="action_message_edit" />

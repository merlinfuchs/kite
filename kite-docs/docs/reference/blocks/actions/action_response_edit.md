---
sidebar_position: 5
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Edit response message

<EmbedFlowNode type="action_response_edit" />

The `Edit response message` block is used to edit a previously created response message. Right now you can only edit the original (first) response message.

You can either configure the message in the block directly or use a message template instead. In both cases you can add embeds and interactive components to the message. The only case where it's better to use a message template is when you want to use the same response in multiple places.

If the message contains interactive components, the flow will be suspended until the user interacts with the message. See [Sub-Flows](/reference/sub-flows) for more information on how interactive components work.

<NodeInfoExplorer type="action_response_edit" />

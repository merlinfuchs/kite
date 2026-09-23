---
sidebar_position: 4
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Create response message

<EmbedFlowNode type="action_response_create" />

The `Create response message` block is used to respond to a command or component interaction. Usually each command or component interaction must create a message response. The only exception is when you show a modal instead.

You can either configure the message in the block directly or use a message template instead. In both cases you can add embeds, components v2 layouts, and interactive components to the message. The only case where it's better to use a message template is when you want to use the same response in multiple places.

If the message contains interactive components, the flow will be suspended until the user interacts with the message. See [Sub-Flows](/reference/sub-flows) for more information on how interactive components work.

:::tip
In the blocks attached to a button or select menu, `user` is whoever clicked it. Use `origin.user` for the person who ran the command. See [Who is `user` in a sub-flow?](/reference/sub-flows#who-is-user-in-a-sub-flow).
:::

<NodeInfoExplorer type="action_response_create" />

---
sidebar_position: 12.1
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# List channel messages

<EmbedFlowNode type="action_message_list" />

The `List channel messages` block gets the latest messages of a channel, newest first. Leave the channel empty to use the channel the flow runs in.

The result is a list of messages, so `{{ result('list')[0].content }}` is the content of the newest one. Use a loop block to go through all of them.

<NodeInfoExplorer type="action_message_list" />

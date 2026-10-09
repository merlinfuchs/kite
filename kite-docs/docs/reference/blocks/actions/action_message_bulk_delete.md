---
sidebar_position: 11.1
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Bulk delete messages

<EmbedFlowNode type="action_message_bulk_delete" />

The `Bulk delete messages` block deletes 2 to 100 messages of a channel at once. Pass the message IDs separated by commas, or a placeholder with a list of messages, like the result of the [List channel messages](./action_message_list.md) block.

Discord doesn't delete messages that are older than 2 weeks this way, and fails the whole request if one of them is.

<NodeInfoExplorer type="action_message_bulk_delete" />

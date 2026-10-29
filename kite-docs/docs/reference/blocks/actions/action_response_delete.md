---
sidebar_position: 8
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Delete response message

<EmbedFlowNode type="action_response_delete" />

The `Delete response message` block is used to delete a previously created response message. By default it deletes the original (first) response. Pick `Other Response` and enter its ID, like `{{result('response_xxx').id}}` of another `Create response message` block, to delete a later one. It must be a response to the same interaction.

<NodeInfoExplorer type="action_response_delete" />

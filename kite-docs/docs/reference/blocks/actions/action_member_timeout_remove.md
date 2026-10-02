---
sidebar_position: 43
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Remove member timeout

<EmbedFlowNode type="action_member_timeout_remove" />

The `Remove member timeout` block lifts the timeout of a member, so they can talk in the server again before the timeout would have ended.

It does nothing if the member isn't timed out. The app needs the **Moderate Members** permission.

<NodeInfoExplorer type="action_member_timeout_remove" />

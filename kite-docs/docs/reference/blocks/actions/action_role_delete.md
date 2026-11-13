---
sidebar_position: 17.3
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Delete role

<EmbedFlowNode type="action_role_delete" />

The `Delete role` block deletes a role from the server the flow runs in, or from another server by its ID. Members who had the role lose it.

The app can only delete roles that are below its own highest role.

<NodeInfoExplorer type="action_role_delete" />

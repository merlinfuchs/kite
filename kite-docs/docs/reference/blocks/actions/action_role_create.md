---
sidebar_position: 17.1
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Create role

<EmbedFlowNode type="action_role_create" />

The `Create role` block creates a new role in the server the flow runs in, or in another server by its ID.

The color can be picked with the color picker, or set with a placeholder that gives a hex code like `#ff0000` or a number like `16711680`.

The result is the created role, so later blocks can use e.g. `{{ result('role').id }}` to give it to a member.

<NodeInfoExplorer type="action_role_create" />

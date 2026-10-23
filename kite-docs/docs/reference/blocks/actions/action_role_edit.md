---
sidebar_position: 17.2
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Edit role

<EmbedFlowNode type="action_role_edit" />

The `Edit role` block changes the name, permissions, color, or other settings of a role in the server the flow runs in, or in another server by its ID.

Settings you leave empty keep what the role already has, so you only need to fill in what should change.

The color can be picked with the color picker, or set with a placeholder that gives a hex code like `#ff0000` or a number like `16711680`.

The result is the edited role, so later blocks can use e.g. `{{ result('role').name }}`.

<NodeInfoExplorer type="action_role_edit" />

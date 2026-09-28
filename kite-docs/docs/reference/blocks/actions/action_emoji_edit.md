---
sidebar_position: 45
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Edit emoji

<EmbedFlowNode type="action_emoji_edit" />

The `Edit emoji` block renames a custom emoji in a server.

The target can be the emoji's ID, the emoji itself like `<:name:123>`, or the result of a `Create emoji` block. The result is the updated emoji. The app needs the **Manage Expressions** permission.

<NodeInfoExplorer type="action_emoji_edit" />

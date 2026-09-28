---
sidebar_position: 48
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Edit sticker

<EmbedFlowNode type="action_sticker_edit" />

The `Edit sticker` block changes the name, emoji or description of a sticker in a server. Fields left empty keep their current value.

The target can be the sticker's ID or the result of a `Create sticker` block. The result is the updated sticker. The app needs the **Manage Expressions** permission.

<NodeInfoExplorer type="action_sticker_edit" />

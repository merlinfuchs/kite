---
sidebar_position: 27
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Leave server

<EmbedFlowNode type="action_server_leave" />

The `Leave Server` block makes the app leave a server by its ID. The app is removed from the server immediately and has to be invited again to rejoin.

:::warning
Leaving a server can't be undone from inside Kite. Double-check that the server ID points to the server you mean, especially when it comes from a placeholder.
:::

### Options

> `Server` The ID of the server to leave. For the server the flow runs in, use `{{guild.id}}`.

<NodeInfoExplorer type="action_server_leave" />

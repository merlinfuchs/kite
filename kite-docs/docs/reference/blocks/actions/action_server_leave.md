---
sidebar_position: 27
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Leave server

<EmbedFlowNode type="action_server_leave" />

The `Leave Server` block makes the app leave the server the flow runs in. The app is removed from the server immediately and has to be invited again to rejoin.

For example, you can use it in an event listener to leave servers the app shouldn't be in.

The block always leaves the server the flow is running in, so a flow can't make the app leave any other server. It fails in flows that don't run in a server, like commands used in DMs and scheduled event listeners.

:::warning
Leaving a server can't be undone from inside Kite. Make sure the block only runs when you want the app to leave, for example by putting it behind a condition.
:::

<NodeInfoExplorer type="action_server_leave" />

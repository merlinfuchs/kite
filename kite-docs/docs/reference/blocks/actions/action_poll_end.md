---
sidebar_position: 15.4
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# End poll

<EmbedFlowNode type="action_poll_end" />

The `End poll` block immediately ends an active poll sent by the bot. The updated poll message is stored as the result of the block.

If you leave `Channel` empty, Kite uses the channel the flow runs in.

:::info
Only polls created by the bot itself can be ended with this block.
:::

<NodeInfoExplorer type="action_poll_end" />

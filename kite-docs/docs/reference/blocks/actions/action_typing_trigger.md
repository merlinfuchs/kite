---
sidebar_position: 43
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Show typing

<EmbedFlowNode type="action_typing_trigger" />

The `Show typing` block makes your app show as typing in a channel, which is useful before blocks that take a while, like an AI or HTTP request. Leave the channel empty to use the channel the flow runs in.

The typing indicator disappears after about 10 seconds, or as soon as your app sends a message in the channel. Run the block again to keep it going for longer.

In commands and other interactions, Discord shows your app as thinking once the response is deferred, so this block is mostly useful in event listeners or for channels other than the one the interaction happened in.

<NodeInfoExplorer type="action_typing_trigger" />

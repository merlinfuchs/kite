---
sidebar_position: 41
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Leave voice channel

<EmbedFlowNode type="action_voice_channel_leave" />

The `Leave voice channel` block makes your app leave the voice channel it's in, in the server the flow runs in.

To protect your app's connection to Discord, this block shares a rate limit with the other blocks that change the status or voice state. Your app can run 5 of them at once, then one every 10 seconds. Runs over the limit fail.

<NodeInfoExplorer type="action_voice_channel_leave" />

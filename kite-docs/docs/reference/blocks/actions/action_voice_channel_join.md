---
sidebar_position: 40
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Join voice channel

<EmbedFlowNode type="action_voice_channel_join" />

The `Join voice channel` block makes your app join a voice or stage channel. If the app is already in a voice channel in that server, it moves to the new one.

Kite can't play or record audio, so the app will just sit in the channel. It may also leave the channel when Kite restarts or reconnects to Discord.

To protect your app's connection to Discord, this block shares a rate limit with the other blocks that change the status or voice state. Your app can run 5 of them at once, then one every 10 seconds. Runs over the limit fail.

<NodeInfoExplorer type="action_voice_channel_join" />

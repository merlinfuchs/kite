---
sidebar_position: 43
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Stop audio

<EmbedFlowNode type="action_voice_audio_stop" />

The `Stop audio` block stops the audio your app plays with [Play audio](./action_voice_audio_play.md), in the server the flow runs in. The app stays in the voice channel, so the next audio starts faster. To also leave the channel, use [Leave voice channel](./action_voice_channel_leave.md) instead.

If the app doesn't play anything in the server, the block does nothing.

<NodeInfoExplorer type="action_voice_audio_stop" />

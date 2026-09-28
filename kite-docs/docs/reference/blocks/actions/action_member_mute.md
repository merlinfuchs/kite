---
sidebar_position: 21.2
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Mute member

<EmbedFlowNode type="action_member_mute" />

The `Mute member` block is used to server-mute a member in voice channels. The member stays muted until they are unmuted, even if they switch or leave voice channels.

The member has to be connected to a voice channel for this to work, and the app needs the `Mute Members` permission. Use the [Unmute member](./action_member_unmute.md) block to undo it.

<NodeInfoExplorer type="action_member_mute" />

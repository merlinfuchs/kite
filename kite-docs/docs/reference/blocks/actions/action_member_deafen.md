---
sidebar_position: 21.4
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Deafen member

<EmbedFlowNode type="action_member_deafen" />

The `Deafen member` block is used to server-deafen a member in voice channels. The member stays deafened until they are undeafened, even if they switch or leave voice channels.

The member has to be connected to a voice channel for this to work, and the app needs the `Deafen Members` permission. Use the [Undeafen member](./action_member_undeafen.md) block to undo it.

<NodeInfoExplorer type="action_member_deafen" />

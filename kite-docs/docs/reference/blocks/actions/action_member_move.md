---
sidebar_position: 21.1
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Move member

<EmbedFlowNode type="action_member_move" />

The `Move member` block is used to move a member to a different voice channel in the server.

The member has to be connected to a voice channel for this to work, and the app needs the `Move Members` permission.

### Options

> `User` The ID of the member to move.
>
> `Channel` The ID of the voice channel to move the member to.
>
> `Audit Log Reason` The reason shown in the server's audit log.

<NodeInfoExplorer type="action_member_move" />

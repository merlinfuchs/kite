---
sidebar_position: 26.1
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Create invite

<EmbedFlowNode type="action_invite_create" />

The `Create invite` block creates an invite for a channel. Leave the channel empty to use the channel the flow runs in.

The result is the created invite. Its link is `https://discord.gg/{{ result('invite').code }}`.

<NodeInfoExplorer type="action_invite_create" />

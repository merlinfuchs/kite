---
sidebar_position: 37.05
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Get Bloxlink User

<EmbedFlowNode type="action_bloxlink_user_get" />

The `Get Bloxlink User` block finds the Roblox account a member verified as with the [Bloxlink](https://blox.link) bot. It returns the same Roblox user as [Get Roblox User](./action_roblox_user_get.md), with their name, display name and avatar.

If the member hasn't verified with Bloxlink, the result is empty. Check it with a Comparison Condition block and tell them to verify first.

It needs the [Bloxlink integration](../../integrations.md), enabled with your server's API key from the Bloxlink dashboard. Bloxlink's server keys only work for their own server, so the block looks members up in the server the flow runs in, unless you set another server ID.

<NodeInfoExplorer type="action_bloxlink_user_get" />

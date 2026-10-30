---
sidebar_position: 37.06
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Get Roblox Group Member

<EmbedFlowNode type="action_roblox_group_member_get" />

The `Get Roblox Group Member` block gets the roles a user has in your Roblox group. The result has `role` and `rank`, the name and rank of their highest role, and `roles` with all of them, highest first.

If the user isn't in the group, the result is empty. Check it with a Comparison Condition block.

The user is a Roblox ID or username. To get the Roblox account of a Discord member, use [Get Bloxlink User](./action_bloxlink_user_get.md) first and pass its result.

It needs the [Roblox Open Cloud integration](../../integrations.md#roblox-open-cloud).

<NodeInfoExplorer type="action_roblox_group_member_get" />

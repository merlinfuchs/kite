---
sidebar_position: 37.07
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Change Roblox Group Role

<EmbedFlowNode type="action_roblox_group_member_role_edit" />

The `Change Roblox Group Role` block ranks a member of your Roblox group, or gives or takes one of their roles. The role can be its name, its rank from 1 to 255, or its ID.

- _Set as only role_ replaces all of the member's roles with this one, like ranking them in a group where everyone has one role.
- _Add role_ and _Remove role_ give or take the role and keep the member's other roles, for groups that give members several.

The user has to be in the group already. The result is the member's roles afterwards, like [Get Roblox Group Member](./action_roblox_group_member_get.md) returns them.

Roblox only lets the owner of the API key change roles below their own highest role. Anyone who can run the command can rank people up to that, so only let staff use it, with the command's permissions or a role check.

It needs the [Roblox Open Cloud integration](../../integrations.md#roblox-open-cloud).

<NodeInfoExplorer type="action_roblox_group_member_role_edit" />

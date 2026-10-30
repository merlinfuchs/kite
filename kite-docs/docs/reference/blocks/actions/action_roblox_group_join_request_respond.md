---
sidebar_position: 37.08
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Answer Roblox Join Request

<EmbedFlowNode type="action_roblox_group_join_request_respond" />

The `Answer Roblox Join Request` block accepts or declines a user's request to join your Roblox group. Use it to let members in once they verified, like after [Get Bloxlink User](./action_bloxlink_user_get.md) found their account.

The result is `true` if the user had a pending request. It's `false` if they didn't ask to join, or are already in the group, so tell them to send a request first.

It needs the [Roblox Open Cloud integration](../../integrations.md#roblox-open-cloud).

<NodeInfoExplorer type="action_roblox_group_join_request_respond" />

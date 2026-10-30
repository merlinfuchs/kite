---
sidebar_position: 37
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Get Roblox User

<EmbedFlowNode type="action_roblox_user_get" />

The `Get Roblox User` block allows you to retrieve information about a Roblox user by their ID or username. This is useful for creating integrations with Roblox or building features that work with Roblox user data.

You can look up users by their Roblox ID or username, and the block will return information about the user. If no user has that ID or username, the result is empty, so check it with a Comparison Condition block before using it.

The result includes `avatar_url`, a picture of the user's avatar head. Use it as the thumbnail of an embed. It's empty if Roblox has blocked the avatar, is still rendering it, or failed to return it. The link stops working after about 30 days, so look the user up again instead of storing it in a variable.

The block needs a Roblox ID or username. A Discord user, like `{{user}}`, doesn't work: Roblox doesn't know which Discord account belongs to which Roblox account. To find a member's Roblox account, use [Get Bloxlink User](./action_bloxlink_user_get.md) if your server verifies members with Bloxlink, or ask for their Roblox username in a command option.

<NodeInfoExplorer type="action_roblox_user_get" />

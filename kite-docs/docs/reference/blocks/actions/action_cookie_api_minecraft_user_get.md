---
sidebar_position: 37.4
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Verify Minecraft player

<EmbedFlowNode type="action_cookie_api_minecraft_user_get" />

The `Verify Minecraft player` block links a Discord user to their Minecraft account with [Cookie API](https://cookie-api.com). Enable Cookie API under [Integrations](../../integrations.md) first.

The player joins `verify.cookie-api.com` in Minecraft, Java or Bedrock, and gets a code. They enter it in a command or modal, and the block returns the player it belongs to, like `{{ result('player').player_name }}`. The block fails if the code is wrong.

<NodeInfoExplorer type="action_cookie_api_minecraft_user_get" />

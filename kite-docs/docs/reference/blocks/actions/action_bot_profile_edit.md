---
sidebar_position: 42.1
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Edit bot server profile

<EmbedFlowNode type="action_bot_profile_edit" />

The `Edit bot server profile` block changes how the bot looks in one server: its display name, bio, avatar and banner. Other servers and the bot's global profile stay as they are. Leave the server empty to use the server the flow runs in.

Settings you leave empty aren't changed, so you can set just the name or just the avatar.

The avatar and banner are URLs of PNG, JPEG, GIF or WebP images up to 4 MB, like `{{ user.avatar_url }}` or the URL of an attachment. Kite downloads the image and uploads it to Discord.

Discord only lets a bot change its profile a few times in a short period, so don't run this block on frequent events.

<NodeInfoExplorer type="action_bot_profile_edit" />

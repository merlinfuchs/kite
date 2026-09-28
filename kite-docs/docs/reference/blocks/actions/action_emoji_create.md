---
sidebar_position: 44
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Create emoji

<EmbedFlowNode type="action_emoji_create" />

The `Create emoji` block uploads a custom emoji to a server.

The image can be uploaded in the editor, or be a URL or placeholder that resolves to one, like `{{arg('image').url}}` for an attachment argument. Emojis can be PNG, JPEG, GIF or WebP and at most 256 KB; 128x128 works best.

Names are 2 to 32 letters, numbers and underscores. Spaces and dashes are turned into underscores.

The result has the emoji's `id`, `name`, `animated`, `url` and `mention`, the emoji as text like `<:name:123>` that you can put in messages. The app needs the **Create Expressions** permission.

<NodeInfoExplorer type="action_emoji_create" />

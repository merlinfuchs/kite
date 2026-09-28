---
sidebar_position: 47
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Create sticker

<EmbedFlowNode type="action_sticker_create" />

The `Create sticker` block uploads a sticker to a server.

A sticker needs a name (2 to 30 characters), an emoji Discord suggests it for, and an image. The description is optional. The image can be uploaded in the editor, or be a URL or placeholder that resolves to one. Stickers can be PNG, APNG or GIF and at most 512 KB; 320x320 works best.

The result has the sticker's `id`, `name`, `description`, `tags` and `url`. The app needs the **Create Expressions** permission.

<NodeInfoExplorer type="action_sticker_create" />

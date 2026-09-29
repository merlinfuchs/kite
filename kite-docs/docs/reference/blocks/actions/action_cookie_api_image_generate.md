---
sidebar_position: 40.2
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Generate AI image

<EmbedFlowNode type="action_cookie_api_image_generate" />

The `Generate AI image` block creates an image from a description with [Cookie API](https://cookie-api.com). Connect Cookie API under [Integrations](../../integrations.md) first.

The result is the image's URL, like `{{ result('image').url }}`. Cookie API keeps generated images for 7 days.

<NodeInfoExplorer type="action_cookie_api_image_generate" />

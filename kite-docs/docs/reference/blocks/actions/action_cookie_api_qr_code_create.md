---
sidebar_position: 40.1
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Generate QR code

<EmbedFlowNode type="action_cookie_api_qr_code_create" />

The `Generate QR code` block creates an image of a QR code with [Cookie API](https://cookie-api.com), for example for a link. Connect Cookie API under [Integrations](../../integrations.md) first.

The result is the image's URL, so `{{ result('qr').url }}` can be the image of an embed.

<NodeInfoExplorer type="action_cookie_api_qr_code_create" />

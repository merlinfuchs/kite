---
sidebar_position: 35
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Send API Request

<EmbedFlowNode type="action_http_request" />

The `Send API Request` block allows you to make HTTP requests to external APIs or web services. You can send GET, POST, PUT, DELETE, and other HTTP methods with custom headers and body data.

This is useful for integrating with external services, fetching data from APIs, or triggering webhooks.

Don't paste API keys into the block. Store them as [secrets](../../secrets.md) and use them as `{{secrets.NAME}}`, so they aren't part of the flow when you share or export it.

<NodeInfoExplorer type="action_http_request" />

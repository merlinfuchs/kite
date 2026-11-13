---
sidebar_position: 7.1
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Respond to webhook

<EmbedFlowNode type="action_webhook_response_create" />

The `Respond to webhook` block answers the request that ran a [webhook event listener](../../event.md#webhook-event-listeners) with a status code and a body. Use it to return data to the service that called the webhook, or to reject a request, for example with status `401` when its signature is invalid.

The status code must be between 200 and 299 or between 400 and 599. The body can be text or JSON and up to 64 KB. A JSON body has to be valid JSON, so check that placeholders in it don't contain quotes.

A flow can respond once, within 10 seconds of the request, and before any wait that ends the run. Otherwise the request is answered with status `202`, and the block only logs a warning.

As long as the flow has this block, every request waits for it, even when the block is in a branch that doesn't run. Respond in every branch, e.g. with status `202`, so the sender doesn't wait for the rest of the flow.

<NodeInfoExplorer type="action_webhook_response_create" />

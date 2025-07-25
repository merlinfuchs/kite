---
sidebar_position: 8
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";

# Defer response

If you know that your flow takes longer than 3 seconds before it can respond, you can use the `Defer response` block to let Discord know that the response is taking longer. After that you have up to 15 minutes to respond.

This block is usually not necessary and can be omitted. Kite is smart enough to detect when your flow is taking too long and will defer the response automatically.

<EmbedFlowNode type="action_response_defer" />

---
sidebar_position: 43.5
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# End Flow

<EmbedFlowNode type="control_flow_end" />

The `End Flow` block stops the flow. No blocks run after it, and loops it's in stop too. Branches of the same block run one after another in the order they were connected, so End Flow also stops the branches connected after its own, but not the ones that already ran.

Ending the flow isn't an error, so an [Error Handler](./control_error_handler.md) doesn't catch it. Commands still need a response, so put a response block before End Flow. A [Wait](./control_sleep.md) longer than 5 seconds that already started in another branch still continues when its time is up.

<NodeInfoExplorer type="control_flow_end" />

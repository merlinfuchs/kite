---
sidebar_position: 43.5
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# End Flow

<EmbedFlowNode type="control_flow_end" />

The `End Flow` block stops the flow. No blocks run after it, including other branches that haven't run yet and loops it's in. Put it at the end of a condition branch to skip the rest of the flow, for example when a user isn't allowed to use a command.

Ending the flow isn't an error, so an [Error Handler](./control_error_handler.md) doesn't catch it. Commands still need a response, so put a response block before End Flow.

<NodeInfoExplorer type="control_flow_end" />

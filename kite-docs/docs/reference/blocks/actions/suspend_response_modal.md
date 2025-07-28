---
sidebar_position: 7
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Show Modal

<EmbedFlowNode type="suspend_response_modal" />

Instead of creating a message response you can also show a modal to the user to ask for further information. Modals can have a number of inputs which you can then access using the `input(...)` variables once the modal has been submitted.

Responding with a modal starts a sub-flow which is suspended until the user submits the modal. See [Sub-Flows](/reference/sub-flows) for more information on how modals work.

<NodeInfoExplorer type="suspend_response_modal" />

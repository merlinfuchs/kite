---
sidebar_position: 45
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Return

<EmbedFlowNode type="control_function_return" />

The `Return` block ends a [function](../../function.md) and returns a value to the `Call function` block that called it. Nothing after it runs in the function. It can only be used in functions.

### Settings

> `Return Value` The value to return, which can contain placeholders. Leave it empty to return nothing.
>
> `Stop Calling Flow` Also stop the flow that called the function, so nothing after its `Call function` block runs. If functions call each other, every one of them stops, up to the command or event listener that started it.

<NodeInfoExplorer type="control_function_return" />

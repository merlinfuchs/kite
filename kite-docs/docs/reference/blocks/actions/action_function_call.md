---
sidebar_position: 43
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Call function

<EmbedFlowNode type="action_function_call" />

The `Call function` block runs one of your [functions](../../function.md) and continues with the blocks after it once the function has finished.

### Settings

> `Function` The function to run. Create functions on the Functions page of your app.
>
> `Arguments` The value to pass for each parameter of the function. They can contain placeholders.

### Output

The value the function returned with its `Return` block, or nothing if it didn't return a value. The block itself doesn't cost credits, only the blocks of the function do.

<NodeInfoExplorer type="action_function_call" />

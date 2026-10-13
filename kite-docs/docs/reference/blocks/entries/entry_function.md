---
sidebar_position: 4
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Function

<EmbedFlowNode type="entry_function" />

The `Function` block is the entry point of a [function](../../function.md). The blocks connected to it run when a `Call function` block calls the function.

### Settings

> `Name` The name of the function, shown when you pick it in the `Call function` block.
>
> `Description` What the function does. This is optional.
>
> `Parameters` The values the function receives. Read them in the function with `{{arg('name')}}`.

<NodeInfoExplorer type="entry_function" />

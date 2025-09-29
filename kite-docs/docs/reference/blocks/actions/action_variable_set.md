---
sidebar_position: 30
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Set stored variable

<EmbedFlowNode type="action_variable_set" />

The `Set stored variable` block is used to store a value in a variable that can be retrieved later using the `Get stored variable` block.

### Settings

> `Variable` Select a variable from your [Stored Variables](https://docs.kite.onl/reference/variable) list.
> 
> `Operation` What should your variable do?
>
> `Value` The value thats changed.

### Output
If you want to use your variable you will use a `Get Variable` block.

<NodeInfoExplorer type="action_variable_set" />

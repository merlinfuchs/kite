---
sidebar_position: 40.5
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Category Condition

<EmbedFlowNode type="control_condition_category" />

The `Category Condition` block runs different actions depending on which category a channel is in. Each `Match Category` branch checks one category, so a single branch covers every channel under that category, including channels you add to it later.

Leave **Channel** empty to check the channel the command or event runs in, or set it to another channel, like a channel command option or the result of an earlier block.

Each branch has two modes:

- **In Category**: runs if the channel is in the category.
- **Not In Category**: runs if the channel isn't in the category.

Add an `Else` branch to run actions when no branch matches, for example to tell the user the command only works in a certain category.

A few things to keep in mind:

- A thread counts as being in the category of the channel it was created in.
- A channel that isn't in any category never matches **In Category**, and always matches **Not In Category**.
- If **Channel** is the ID of a category itself, it counts as being in that category.

To check a single channel instead of a whole category, use the [Channel Condition](./control_condition_channel.md) block.

<NodeInfoExplorer type="control_condition_category" />

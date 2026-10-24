---
sidebar_position: 33
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Generate random value

<EmbedFlowNode type="action_random_generate" />

The `Generate random value` block is used to generate a random number within a specified range, or to pick a random item from a list.

To pick an item, set the mode to `Item from list` and put one item per line. Empty lines are ignored. The items can also be a single placeholder of a list, like `{{var('prizes')}}` or the result of a block that returns a list. Its items keep their type, so a picked member can be mentioned with `{{result('id').mention}}`.

<NodeInfoExplorer type="action_random_generate" />

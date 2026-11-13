---
sidebar_position: 33
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Generate random value

<EmbedFlowNode type="action_random_generate" />

The `Generate random value` block is used to generate a random number within a specified range, or to pick random items from a list.

To pick an item, set the mode to `Item from list` and put one item per line. Empty lines are ignored. The items can also be a single placeholder of a list, like `{{var('prizes')}}` or the result of a block that returns a list. Its items keep their type, so a picked member can be mentioned with `{{result('id').mention}}`.

To pick several items, like the winners of a giveaway, set the mode to `Several items from list` and the count to how many to pick, from 1 to 100. The same entry of the list is never picked twice, so put each item in the list once. The result is a list of the picked items in random order. If the list has fewer items than the count, all of them are picked, and none from an empty list. Check with `{{len(result('id')) == 0}}` whether anything was picked.

<NodeInfoExplorer type="action_random_generate" />

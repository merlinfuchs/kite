---
sidebar_position: 37.5
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# If

<EmbedFlowNode type="control_if" />

The `If` block runs the blocks after its `Yes` output if a condition is true, and the blocks after `No` if it isn't.

Set the condition to a single placeholder that is true or false, with the whole comparison inside the brackets, like `{{message.content != ''}}`, `{{result('count') > 5}}` or `{{event.emoji.name == '⭐'}}`. The `Yes` branch also runs for `true`, `yes` and `1` in any case, like the value of a variable or a select menu option. Anything else runs `No`, including `no`, `0`, `2` and an empty value.

For more than two branches, or to compare text without writing an expression, use a [Comparison Condition](./control_condition_compare.md).

<NodeInfoExplorer type="control_if" />

---
sidebar_position: 37.5
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# If

<EmbedFlowNode type="control_if" />

The `If` block runs the blocks after its `Yes` output if a condition is true, and the blocks after `No` if it isn't.

Set the condition to a placeholder that is true or false, like `{{message.content != ''}}`, `{{result('count') > 5}}` or `{{event.emoji.name == '⭐'}}`. The `Yes` branch also runs for the text `true` in any case, like the value of a variable. Anything else runs `No`, including `yes`, `1` and an empty value.

For more than two branches, or to compare text without writing an expression, use a [Comparison Condition](./control_condition_compare.md).

<NodeInfoExplorer type="control_if" />

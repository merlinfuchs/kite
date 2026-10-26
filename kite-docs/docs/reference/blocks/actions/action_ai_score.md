---
sidebar_position: 32.5
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# AI Score

<EmbedFlowNode type="action_ai_score" />

The `AI Score` block lets AI rate something on a scale you define, like how urgent a ticket is or how toxic a message is. Ask a question, like `How urgent is this ticket?`, set the input to what the AI rates, like `{{message.content}}`, and describe 2 to 10 levels, lowest first.

The block returns the level, from 1 for the first to the number of levels. It can be in between two levels, like `2.4` when the AI isn't sure whether it's the second or the third. Use a [Comparison Condition](../controls/control_condition_compare.md) to act on it, for example to ping a moderator when the result is greater than or equal to `2.5`.

The question, input and levels can be up to 8,000 bytes together, after placeholders are filled in. That's 8,000 characters of English text, and fewer in other scripts or with emoji, which take up to 4 bytes each. The block costs [1 credit](../../credit-system.md#cost-breakdown).

<NodeInfoExplorer type="action_ai_score" />

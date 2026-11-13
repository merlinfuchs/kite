---
sidebar_position: 38.7
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# AI Check

<EmbedFlowNode type="control_ai_check" />

The `AI Check` block lets AI answer a yes or no question, like `Is this message spam?`, and continues with its **Yes**, **No** or **Unsure** branch. Set the input to what the AI decides about, like `{{message.content}}`.

The AI answers with how likely yes is, from 0 to 1, and the thresholds decide which branch runs:

- **Yes** runs when it's at least the yes threshold.
- **No** runs when it's below the no threshold.
- **Unsure** runs in between.

A threshold you leave empty is the same as the other one, and both default to `0.5`. So until you set both, the block only runs **Yes** or **No**. To delete spam only when the AI is sure, set the yes threshold to `0.9` and the no threshold to `0.5`, and send messages in between to your moderators from the **Unsure** branch. The block returns how likely yes is.

The question and input can be up to 8,000 bytes together, after placeholders are filled in. That's 8,000 characters of English text, and fewer in other scripts or with emoji, which take up to 4 bytes each. The block costs [1 credit](../../credit-system.md#cost-breakdown). If the AI can't be reached, the flow fails, which you can catch with a [Handle Errors](./control_error_handler.md) block.

<NodeInfoExplorer type="control_ai_check" />

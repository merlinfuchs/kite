---
sidebar_position: 38.6
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# AI Choice

<EmbedFlowNode type="control_condition_ai_choice" />

The `AI Choice` block lets AI pick one of the options you add and runs the actions of that option. Use it to sort things into categories, like sending a support question to the right channel or telling a bug report from a feature request.

Ask a question, like `What is this ticket about?`, and set the input to what the AI decides about, like `{{message.content}}`. Give each option a name and, if the name doesn't say enough, a description of when to pick it.

The AI also says how confident it is, from 0 to 1. With a **Min Confidence** like `0.8`, the **Else** branch runs instead when the AI is less confident than that. Without one, the AI always picks an option. The block returns the name of the option it picked, also when it wasn't confident enough.

The question, input and options can be up to 8,000 bytes together, after placeholders are filled in. That's 8,000 characters of English text, and fewer in other scripts or with emoji, which take up to 4 bytes each. The block costs [1 credit](../../credit-system.md#cost-breakdown). If the AI can't be reached, the flow fails, which you can catch with a [Handle Errors](./control_error_handler.md) block.

<NodeInfoExplorer type="control_condition_ai_choice" />

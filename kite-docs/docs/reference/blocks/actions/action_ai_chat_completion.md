---
sidebar_position: 31
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Ask AI

<EmbedFlowNode type="action_ai_chat_completion" />

The `Ask AI` block allows you to interact with artificial intelligence models. You can ask questions, get responses to prompts, or have the AI perform various text-based tasks.

You can pick between three model tiers, Fast, Balanced and Smartest, which differ in capability and [credit cost](../../credit-system.md#cost-breakdown). We move each tier to newer models over time, so a flow keeps working without you having to pick a new model. The response can be used in subsequent blocks.

The prompt and system prompt can be up to 64,000 bytes together, after placeholders are filled in, or 8,000 bytes with Smartest. That's 64,000 characters of English text, and fewer in other scripts or with emoji, which take up to 4 bytes each.

The answer can be up to 500 tokens long, about 2,000 characters of English. Set `Max Answer Length` to allow up to 1,000. With Balanced, the AI's reasoning counts toward the length too.

The block's credits cover 8,000 bytes of prompt and 500 tokens of answer. Each further 8,000 bytes of prompt, and an answer over 500 tokens, cost the block's credits once more. A Fast block with a 20,000 byte prompt and a 700 token answer costs 5 + 10 + 5 = 20 credits. See the [credit costs](../../credit-system.md#cost-breakdown).

<NodeInfoExplorer type="action_ai_chat_completion" />

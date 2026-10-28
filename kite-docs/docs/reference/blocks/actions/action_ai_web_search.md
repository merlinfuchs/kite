---
sidebar_position: 32
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Search the Web

<EmbedFlowNode type="action_ai_web_search" />

The `Search the Web` block allows you to search the internet for the latest information using AI. This is useful for getting current information that might not be available in the AI's training data.

This block uses AI to process and summarize web search results, providing you with relevant and up-to-date information.

The prompt can be up to 64,000 bytes, after placeholders are filled in, or 8,000 bytes with Smartest. That's 64,000 characters of English text, and fewer in other scripts or with emoji, which take up to 4 bytes each.

The answer can be up to 500 tokens long, about 2,000 characters of English. Set `Max Answer Length` to allow up to 1,000.

The block's credits cover 8,000 bytes of prompt and 500 tokens of answer. Each further 8,000 bytes of prompt, and an answer over 500 tokens, cost what `Ask AI` costs with the same model once more. See the [credit costs](../../credit-system.md#cost-breakdown).

<NodeInfoExplorer type="action_ai_web_search" />

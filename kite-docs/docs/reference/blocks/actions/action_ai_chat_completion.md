---
sidebar_position: 31
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Ask AI

<EmbedFlowNode type="action_ai_chat_completion" />

The `Ask AI` block allows you to interact with artificial intelligence models. You can ask questions, get responses to prompts, or have the AI perform various text-based tasks.

You can pick between three model tiers, Fast, Balanced and Smartest, which differ in capability and [credit cost](../../credit-system.md#cost-breakdown). We move each tier to newer models over time, so a flow keeps working without you having to pick a new model. The response can be used in subsequent blocks.

The prompt and system prompt can be up to 8,000 bytes together, after placeholders are filled in. That's 8,000 characters of English text, and fewer in other scripts or with emoji, which take up to 4 bytes each.

<NodeInfoExplorer type="action_ai_chat_completion" />

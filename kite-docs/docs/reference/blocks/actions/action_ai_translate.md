---
sidebar_position: 32.6
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# AI Translate

<EmbedFlowNode type="action_ai_translate" />

The `AI Translate` block lets AI translate text to another language. Set the text to what you want translated, like `{{message.content}}`, and pick the language from the list. Mentions, emoji, links and Markdown stay as they are.

To pick the language when the flow runs, choose `From a placeholder` and enter a placeholder that becomes the code of a language, like `es` or `pt-BR`. The [Lookup Value](./action_value_lookup.md) block can turn a flag reaction into a code, which is how the Flag Translator template works.

The block returns the translated text as `text`, and the code of the language the text was in as `source_language`, or `other` if it's none of the languages in the list. Use it to skip messages that are already in the right language.

The text can be up to 4,000 characters, after placeholders are filled in. The block costs [1 credit](../../credit-system.md#cost-breakdown) for every 1,000 characters, so a message of 2,500 characters costs 3 credits.

<NodeInfoExplorer type="action_ai_translate" />

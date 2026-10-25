---
sidebar_position: 37.82
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Transform text with Popcat

<EmbedFlowNode type="action_popcat_text_transform" />

The `Transform text with Popcat` block changes text, like into mOcKiNg text, Morse code or binary. Enable Popcat under [Integrations](../../integrations.md) first. It needs no key, but it's off until you turn it on, as its blocks send what you enter to [Popcat](https://popcat.xyz/api), a free API that isn't run by Kite.

Pick the transformation and enter the text, which can be a placeholder like `{{ arg('text') }}`. Most return the new text as `text`, so `{{ result('text').text }}`, and Morse code as `morse`.

| Transform | Settings | What it returns |
| --- | --- | --- |
| Mocking text | Text | Alternates upper and lower case, like mOcKiNg TeXt. Returns text. |
| Reverse | Text | Reverses the text. Returns text. |
| Double-struck | Text | Writes the text in double-struck letters, like 𝕥𝕙𝕚𝕤. Returns text. |
| LOLcat | Text | Translates the text into LOLcat speak. Returns text. |
| Text to Morse code | Text | Translates the text into Morse code. Returns morse. |
| Text to binary | Text | Encodes the text as binary. Returns text. |
| Binary to text | Binary | Decodes binary back into text. Returns text. |

<NodeInfoExplorer type="action_popcat_text_transform" />

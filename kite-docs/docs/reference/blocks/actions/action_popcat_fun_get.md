---
sidebar_position: 37.81
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Get Popcat fun

<EmbedFlowNode type="action_popcat_fun_get" />

The `Get Popcat fun` block gets something random from Popcat, like a joke, a fact or a magic 8-ball answer. Enable Popcat under [Integrations](../../integrations.md) first. It needs no key, but it's off until you turn it on, as its blocks send what you enter to [Popcat](https://popcat.xyz/api), a free API that isn't run by Kite.

Pick what to get in the block. The result is what Popcat returned, so use the field listed for it, like `{{ result('fun').answer }}` for the magic 8-ball or `{{ result('fun').ops1 }}` and `{{ result('fun').ops2 }}` for the two options of a would you rather question.

| Get | Settings | What it returns |
| --- | --- | --- |
| Magic 8-ball | - | A random magic 8-ball answer. Returns answer. |
| Joke | - | A random joke. Returns joke. |
| Fact | - | A random fact. Returns fact. |
| Shower thought | - | A random shower thought from Reddit. Returns result, author and upvotes. |
| Would you rather | - | A random would you rather question. Returns its two options as ops1 and ops2. |
| Meme | - | A random meme from Reddit. Returns title, link, subreddit, content.image, misc.upvotes and misc.nsfw, which is true for memes that aren't safe for work. |
| Car | - | A random picture of a car. Returns title and image. |
| Random color | - | A random color. Returns hex, name and image. |

<NodeInfoExplorer type="action_popcat_fun_get" />

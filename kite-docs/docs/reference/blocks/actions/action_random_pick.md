---
sidebar_position: 33.1
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Pick random item

<EmbedFlowNode type="action_random_pick" />

The `Pick random item` block is used to pick a random item from a list. Put one item per line, empty lines are ignored. Use the picked item in later blocks with `{{result('id')}}`, which the placeholder explorer fills in for you.

<NodeInfoExplorer type="action_random_pick" />

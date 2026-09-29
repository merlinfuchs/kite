---
sidebar_position: 41.2
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Delete soundboard sound

<EmbedFlowNode type="action_soundboard_sound_delete" />

The `Delete soundboard sound` block is used to remove a sound from a server's soundboard by its ID. The app needs the `Create Expressions` permission to delete sounds it created, and `Manage Expressions` to delete sounds created by others.

### Options

> `Target Sound` The ID of the soundboard sound to delete.

<NodeInfoExplorer type="action_soundboard_sound_delete" />

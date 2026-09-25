---
sidebar_position: 42
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Set status

<EmbedFlowNode type="action_status_set" />

The `Set status` block changes the status and activity your app shows in Discord. The activity name supports placeholders, so you can show things like the number of members in a server. This block requires Premium.

The status only lasts until something else changes it. Saving the status in the app settings, Kite reconnecting to Discord, or the next step of a rotating status all replace it with the status from the app settings.

To protect your app's connection to Discord, this block shares a rate limit with the other blocks that change the status or voice state. Your app can run 5 of them at once, then one every 10 seconds. Runs over the limit fail.

<NodeInfoExplorer type="action_status_set" />

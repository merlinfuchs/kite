---
sidebar_position: 4
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";

# Bot Left Server

<EmbedFlowNode type="entry_event" />

Triggers a flow when the app is removed from a server. Select `Bot Left Server` as the event type on the `Listen for Event` block.

It fires when the app is kicked, banned, or the server is deleted, but not during a temporary Discord outage. The server is available as `{{guild.id}}`.

Example: log which servers remove the app so you can track churn.

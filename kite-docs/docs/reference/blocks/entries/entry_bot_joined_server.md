---
sidebar_position: 3
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";

# Bot Joined Server

<EmbedFlowNode type="entry_event" />

Triggers a flow when the app is added to a server. Select `Bot Joined Server` as the event type on the `Listen for Event` block.

It only fires for this app, never for other bots, and only on a genuine join, not when the app reconnects to servers it is already in. The server is available as `{{guild.id}}`.

Example: send a welcome message to the server owner when the app is added to a new server.

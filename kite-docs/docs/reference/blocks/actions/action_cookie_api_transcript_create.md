---
sidebar_position: 37.1
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Create transcript

<EmbedFlowNode type="action_cookie_api_transcript_create" />

The `Create transcript` block saves the last 500 messages of a channel as a web page with [Cookie API](https://cookie-api.com), for example before a ticket is closed. Enable Cookie API under [Integrations](../../integrations.md) first.

Cookie API reads the messages itself, so the block sends it your bot's token. You don't need to paste the token anywhere, and it isn't part of your flow. Running the block again for the same channel updates its transcript.

The result is the link, like `{{ result('transcript').url }}`, which you can send to the ticket's creator or a log channel.

<NodeInfoExplorer type="action_cookie_api_transcript_create" />

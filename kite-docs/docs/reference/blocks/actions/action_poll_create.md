---
sidebar_position: 15.3
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Create poll

<EmbedFlowNode type="action_poll_create" />

The `Create poll` block is used to send a poll to a specific channel. The created message is stored as the result of the block, so later blocks can use its ID, for example to pin it.

A poll has a question of up to 300 characters and between 1 and 10 answers of up to 55 characters each. Every answer can have an emoji next to it. The question and the answers can contain placeholders. Answers that are empty after the placeholders are filled in are skipped, so you can use optional command arguments as answers.

The poll stays open for 24 hours unless you set a duration between 1 and 768 hours (32 days). Turn on `Allow Multiple Answers` to let people vote for more than one answer.

:::info
The bot needs the `Send Polls` permission in the channel. Polls can't be sent to announcement channels.
:::

<NodeInfoExplorer type="action_poll_create" />

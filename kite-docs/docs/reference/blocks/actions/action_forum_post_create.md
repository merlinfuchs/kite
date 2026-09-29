---
sidebar_position: 43
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Create forum post

<EmbedFlowNode type="action_forum_post_create" />

The `Create forum post` block is meant to create a new post in a forum channel.

:::caution Not available yet

This block isn't finished. It isn't offered in the block list and doesn't create a post when it runs, so flows continue as if it wasn't there. To start a discussion in the meantime, use the [Create thread](./action_thread_create.md) block.

:::

### Options

> `Channel Target` The ID of the forum channel to create the post in.
>
> `Channel Data` The name and other settings of the post.
>
> `Audit Log Reason` This will appear in the Discord audit log.

<NodeInfoExplorer type="action_forum_post_create" />

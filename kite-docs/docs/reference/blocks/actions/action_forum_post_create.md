---
sidebar_position: 26.2
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Create forum post

<EmbedFlowNode type="action_forum_post_create" />

The `Create forum post` block is used to create a post in a forum or media channel. The message you configure becomes the first message of the post, and can have buttons and select menus like any other message.

### Options

> `Channel` The forum or media channel to create the post in.
>
> `Post Title` The title of the post, up to 100 characters.
>
> `Tags` Names of up to 5 of the forum's tags, separated by commas. Leave empty for a post without tags. The block fails if the forum has no tag with one of the names, or if the forum requires a tag and none is set.

<NodeInfoExplorer type="action_forum_post_create" />

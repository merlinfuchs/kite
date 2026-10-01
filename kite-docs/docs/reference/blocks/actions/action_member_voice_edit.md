---
sidebar_position: 41.1
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Edit voice state

<EmbedFlowNode type="action_member_voice_edit" />

The `Edit voice state` block is used to server-mute, server-deafen or move a member in voice channels. You can change any combination of the three in one block.

Server mute and deafen stay in place until they are turned off again, even if the member switches or leaves voice channels. Moving only works while the member is connected to a voice channel. The app needs the `Mute Members`, `Deafen Members` or `Move Members` permission for the changes you make.

### Options

> `User` The ID of the member to edit.
>
> `Server Mute` Turn the member's server mute on or off, or leave it unchanged.
>
> `Server Deafen` Turn the member's server deafen on or off, or leave it unchanged.
>
> `Move to Channel` The ID of the voice channel to move the member to. Leave it empty to keep the member where they are. If it's set but doesn't resolve to a valid channel ID, the block fails instead of skipping the move.
>
> `Audit Log Reason` The reason shown in the server's audit log.

The block fails if nothing is changed, so at least one of the three options has to be set. The editor shows an error until one is.

<NodeInfoExplorer type="action_member_voice_edit" />

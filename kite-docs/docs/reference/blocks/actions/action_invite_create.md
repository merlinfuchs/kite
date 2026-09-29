---
sidebar_position: 26
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Create invite

<EmbedFlowNode type="action_invite_create" />

The `Create invite` block is used to create an invite link for a channel. By default the invite is created for the channel the flow runs in, but you can pick a different channel by its ID.

The app needs the `Create Invite` permission in the target channel.

### Options

> `Channel` The ID of the channel to create the invite for. Leave empty to use the channel the flow runs in.
>
> `Max Age` How many seconds the invite lasts before it expires. Set it to `0` for an invite that never expires, or leave it empty to use Discord's default of 24 hours. Must be a whole number between `0` and `604800` (7 days).
>
> `Max Uses` How many times the invite can be used before it stops working. Leave it empty or set it to `0` for unlimited uses. Must be a whole number between `0` and `100`.
>
> If `Max Age` or `Max Uses` is set but doesn't evaluate to a valid number in range (for example a placeholder that resolves to nothing, text, a decimal or a negative number), the block fails with an error instead of creating the invite.
>
> `Temporary` Members who join through this invite are kicked once they go offline, unless they've been given a role.
>
> `Unique` Always create a new invite instead of possibly reusing a similar unused one.
>
> `Audit Log Reason` The reason shown in the server's audit log.

### Result

The created invite is available to later blocks. Using `{{result('CREATE_INVITE')}}`, where `CREATE_INVITE` is the ID of this block, inserts the full invite URL.

<NodeInfoExplorer type="action_invite_create" />

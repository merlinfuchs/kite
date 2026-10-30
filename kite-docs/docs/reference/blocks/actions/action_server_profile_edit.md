---
sidebar_position: 44
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Edit server profile

<EmbedFlowNode type="action_server_profile_edit" />

The `Edit server profile` block changes your bot's nickname, avatar, banner or About Me in one server, without changing them anywhere else. It uses the server the flow runs in, unless you pick another one. This block requires Premium.

Only the fields you fill in change. The avatar and banner are URLs of a PNG, JPEG, GIF or WebP image up to 10 MB, which Kite downloads, so `{{user.avatar_url}}` gives your bot the avatar of the user who ran the command. A field whose placeholders turn out empty resets that part to your bot's own profile. Turn on `Reset Profile` to reset all of them at once.

You can also edit the server profile on the `Servers` page of your app.

<NodeInfoExplorer type="action_server_profile_edit" />

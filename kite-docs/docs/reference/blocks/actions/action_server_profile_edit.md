---
sidebar_position: 44
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Edit server profile

<EmbedFlowNode type="action_server_profile_edit" />

The `Edit server profile` block changes your bot's nickname, avatar, banner, About Me, or name style in one server, without changing them anywhere else. It uses the server the flow runs in, unless you pick another one. This block needs Kite Premium or Kite Ultimate.

Only the fields you fill in change. The avatar and banner are URLs of a PNG, JPEG, GIF or WebP image up to 10 MB, which Kite downloads, so `{{user.avatar_url}}` gives your bot the avatar of the user who ran the command. The upload button next to them uploads an image to Kite and fills in its URL. A field whose placeholders turn out empty resets that part to your bot's own profile. Turn on `Reset profile` to reset all of them at once.

The name style gives your bot's name a font, an effect, and colors, like Discord Nitro members can. Servers show the font. The bot's profile also shows the effect and colors, and servers show the role color instead. Solid, Neon, Toon, and Pop use one color, Gradient two, Gummy four, and Prism five. Colors can be placeholders that turn into a hex color like `#5865F2`. Set `Name font` back to `Keep current` to leave the style as it is.

The profile can be changed 5 times a minute in each server, the block errors when it's changed more often.

You can also edit the server profile on the `Servers` page of your app.

<NodeInfoExplorer type="action_server_profile_edit" />

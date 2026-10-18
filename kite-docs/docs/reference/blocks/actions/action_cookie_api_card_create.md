---
sidebar_position: 37.2
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Generate card

<EmbedFlowNode type="action_cookie_api_card_create" />

The `Generate card` block creates an image from a design, like a welcome or level card, with [Cookie API](https://cookie-api.com). Enable Cookie API under [Integrations](../../integrations.md) first.

Design the card in the [Card Builder](https://trolensdesign.github.io/CardBuilder-cookie-api/), copy its JSON and paste it into the block. Placeholders work in the texts and URLs of the JSON, so `"text": "Welcome {{ user.display_name }}"` greets every new member by name, and `"user_id": "{{ user.id }}"` shows their avatar in a Discord profile element.

The result is the image's URL, so `{{ result('card').url }}` can be the image of an embed. Cookie API deletes cards after 7 days.

<NodeInfoExplorer type="action_cookie_api_card_create" />

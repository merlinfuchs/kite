---
sidebar_position: 37.4
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Create captcha

<EmbedFlowNode type="action_cookie_api_captcha_create" />

The `Create captcha` block creates a page with a captcha from Cloudflare or Google, hosted by [Cookie API](https://cookie-api.com). Enable Cookie API under [Integrations](../../integrations.md) first.

Use it to verify new members: send them `{{ result('captcha').url }}` in a link button, and save `{{ result('captcha').captcha_id }}` in a variable for the user. A second button then runs [Check captcha](./action_cookie_api_captcha_get.md) with the saved ID, and gives the member a role if they solved it. The captcha expires after an hour.

<NodeInfoExplorer type="action_cookie_api_captcha_create" />

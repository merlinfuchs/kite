---
sidebar_position: 37.5
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Check captcha

<EmbedFlowNode type="action_cookie_api_captcha_get" />

The `Check captcha` block checks whether a user solved a captcha from [Create captcha](./action_cookie_api_captcha_create.md). Enable Cookie API under [Integrations](../../integrations.md) first.

`{{ result('check').solved }}` is `YES` once the user solved it and `NO` before, so a comparison condition can decide what happens next.

<NodeInfoExplorer type="action_cookie_api_captcha_get" />

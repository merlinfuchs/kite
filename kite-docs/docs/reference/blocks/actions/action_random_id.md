---
sidebar_position: 32.5
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Generate random ID

<EmbedFlowNode type="action_random_id" />

The `Generate random ID` block generates a random ID or code, like a ticket number or a verification code.

It can generate these formats:

- **UUID v4**: A 36 character ID like `9b2f6c1e-4d3a-4f8b-a1c2-7e5d9f0b3a64`. It always has the same length.
- **Nano ID**: Letters, numbers, `-` and `_`. 21 characters by default.
- **Alphanumeric**: Letters and numbers. 16 characters by default.
- **Numeric**: Only numbers, for PINs or verification codes. 6 characters by default.
- **Hexadecimal**: Numbers and the letters `a` to `f`. 32 characters by default.

The length can be from 1 to 256 characters, or a placeholder. If a placeholder doesn't resolve to a whole number in that range, the block fails.

The result is the generated ID as text, so later blocks can use e.g. `{{ result('id') }}` to show it in a message.

<NodeInfoExplorer type="action_random_id" />

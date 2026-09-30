---
sidebar_position: 32.5
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Generate Random ID

<EmbedFlowNode type="action_random_id" />

The `Generate Random ID` block generates random strings and identifiers for use across your flow.

### ID Types

You can select from the following identifier formats:

- **UUID v4**: A standard 36-character globally unique identifier formatted as `xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx`. The length is fixed.
- **Nano ID**: A compact, URL-safe identifier containing letters, numbers, hyphens, and underscores. Defaults to 21 characters.
- **Alphanumeric**: A random string of upper- and lower-case letters and numbers (`0-9`, `a-z`, `A-Z`). Defaults to 16 characters.
- **Numeric**: A random string of numbers (`0-9`). Defaults to 6 characters. Useful for verification codes, PINs, or ticket numbers.
- **Hexadecimal**: A random hexadecimal string (`0-9`, `a-f`). Defaults to 32 characters. Useful for color hexes, hashes, and token IDs.

### Options

> `ID Type` The format of identifier to generate.
>
> `Length` The character length of the generated ID (applicable to Nano ID, Alphanumeric, Numeric, and Hexadecimal). You can enter a number or provide a variable placeholder. If left empty, the default length for the selected type is used. Maximum supported length is 256 characters.

### Output

The generated value is stored as a string and can be referenced in downstream blocks:

- **`{id}`**: The generated identifier string.

You can use `{id}` in message content, embed fields, variable stores, channel names, or HTTP requests.

<NodeInfoExplorer type="action_random_id" />

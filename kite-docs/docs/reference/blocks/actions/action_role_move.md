---
sidebar_position: 26
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Move role

<EmbedFlowNode type="action_role_move" />

The `Move role` block changes a role's place in the server's role hierarchy. Creating or editing a role does not set its position, so this is a separate block (it uses its own Discord API call).

### How the hierarchy works

Every role has a `Position`, a whole number. A higher number sits higher in the list and takes precedence for color and permissions. The default `@everyone` role is always `0`.

> `Target Role` The ID of the role to move.
>
> `Position` The new position. `1` is just above `@everyone`; larger numbers move the role further up.

The app can only move a role below its own highest role, and cannot move a role above itself. If the position is out of range, Discord clamps it to the nearest valid spot.

### Example

Move a `Booster` role up when someone boosts the server:

1. `Booster` role added (entry or another block).
2. `Move role` with `Target Role` set to the booster role and `Position` set to, for example, `{{var('boost_position')}}` or a fixed number like `5`.

The block returns the role's ID, which you can store in a temporary variable to use in later blocks.

<NodeInfoExplorer type="action_role_move" />

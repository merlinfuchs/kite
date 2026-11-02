---
sidebar_position: 24
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Get member status

<EmbedFlowNode type="action_member_presence_get" />

The `Get member status` block gets the status and custom status of a member, so a [Comparison Condition](../controls/control_condition_compare.md) can check it. For example, to only let members with `sub to jack` in their status use a command, compare `{{result('presence_abc').custom_status}}` with **Contains** `sub to jack`.

| Placeholder                                      | Value                                                                       |
| ------------------------------------------------ | --------------------------------------------------------------------------- |
| `{{result('presence_abc').status}}`              | `online`, `idle`, `dnd` or `offline`                                        |
| `{{result('presence_abc').custom_status}}`       | The text of the custom status, empty without one                            |
| `{{result('presence_abc').custom_status_emoji}}` | The emoji of the custom status, a unicode emoji or the custom emoji mention |
| `{{result('presence_abc').activity}}`            | What the member is doing, like the game they play, empty if nothing         |

Replace `presence_abc` with the ID of your block.

- Your bot needs the **Presence Intent**. Turn it on under **Bot** in the [Discord Developer Portal](https://discord.dev). Kite requests it for apps that use the block, which can take up to a minute after saving.
- Discord has no way to ask for the status of a member, it sends status changes to bots while they're online. Kite keeps the latest one, so the block costs no credits.
- Invisible members appear offline, like they do to other members. Offline members have no custom status.
- To run a flow when a member's status changes, use the [Presence Update](../../event.md#presence-update) event listener instead.

<NodeInfoExplorer type="action_member_presence_get" />

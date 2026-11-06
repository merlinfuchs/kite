---
sidebar_position: 48
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Event Filter

<EmbedFlowNode type="option_event_filter" />

The `Event Filter` block allows you to filter events based on their properties. You can specify conditions that must be met for the event to trigger your flow.

You can add several filters to the same event listener. The flow only runs when the event matches all of them.

## Filter targets

| Target                 | Value                                                        | Events                                                                                |
| ---------------------- | ------------------------------------------------------------ | ------------------------------------------------------------------------------------- |
| Message content        | Text, compared with the filter mode                          | Message create, Message update, Message delete, Direct message received               |
| User ID                | A user ID                                                    | All except Bot joined server and Bot left server                                      |
| Server ID              | A server ID                                                  | All except Direct message received                                                    |
| Channel ID             | A channel ID                                                 | Messages, reactions, Server boost, Audit log entry                                    |
| Message ID             | A message ID                                                 | Messages, reactions, Server boost                                                     |
| Member role            | A role ID or mention, with `Has role` or `Doesn't have role` | Message create, Message update, Message reaction add, Server member add, Server boost |
| Reaction emoji         | An emoji, picked with the emoji picker                       | Message reaction add, Message reaction remove                                         |
| Message has attachment | Yes or No                                                    | Message create, Message update, Direct message received                               |
| Message mentions bot   | Yes or No, whether the message pings your bot                | Message create, Message update, Direct message received                               |
| User is bot            | Yes or No. Choose No to ignore bots                          | Server member add, Server member remove, Message reaction add                         |
| Audit log action       | The actions, picked from a list                              | Audit log entry                                                                       |
| Voice action           | The actions, picked from a list                              | Voice state update                                                                    |

The editor only offers the targets the event of your listener has. A filter on something the event doesn't have, like the message content of a Member Join, would stop the listener from ever running, so it can't be saved. Messages from bots are always ignored, which is why User is bot isn't offered for message events.

Message Content filters are case-sensitive. Enable `Ignore Case` so that `!help` also matches `!Help`.

**Is One Of** and **Is Not One Of** compare with several values separated by commas, for example a few channel IDs.

Values that aren't picked from a list can use [parameters](../../parameters.md), like `{{params.suggestion_channel}}` for a Channel ID or `{{params.staff_role}}` for a Member role. Other placeholders don't work in filters, because filters run before the flow starts.

For [Audit log entry](../../event.md#audit-log-entry) listeners, filter on the **Audit log action** and pick the actions the flow runs for, like Member Banned and Member Kicked.

On Message delete, the message content and the user ID of the author are only known for messages Kite saw before they were deleted.

Filtering on the message ID and the reaction emoji is useful for reaction roles: the flow only runs for the right emoji on the right message.

## How filters run

Kite checks filters before the flow starts. Events that don't match don't run any blocks, unlike a condition block, which only runs once the flow has started. If a listener only cares about some events, like messages in one channel, a filter is the better choice.

<NodeInfoExplorer type="option_event_filter" />

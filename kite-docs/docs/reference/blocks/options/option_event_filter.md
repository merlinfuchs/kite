---
sidebar_position: 48
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Event Filter

<EmbedFlowNode type="option_event_filter" />

The `Event Filter` block allows you to filter events based on their properties. You can specify conditions that must be met for the event to trigger your flow.

You can add several filters to the same event listener. The flow only runs when the event matches all of them.

## Filter Targets

| Target                 | Value                                                         | Events                                                                  |
| ---------------------- | ------------------------------------------------------------- | ----------------------------------------------------------------------- |
| Message Content        | Text, compared with the filter mode                           | Message Create, Message Update                                          |
| User ID                | A user ID                                                     | All except Message Delete and the bot server events                     |
| Guild ID               | A server ID                                                   | All                                                                     |
| Channel ID             | A channel ID                                                  | Messages, reactions, Server Boost                                       |
| Message ID             | A message ID                                                  | Messages, reactions, Server Boost                                       |
| Member Role            | A role ID or mention, with `Has Role` or `Doesn't Have Role`  | Message Create, Message Update, Reaction Add, Member Join, Server Boost |
| Reaction Emoji         | A unicode emoji, or the ID, name or mention of a custom emoji | Reaction Add, Reaction Remove                                           |
| Message Has Attachment | Yes or No                                                     | Message Create, Message Update                                          |
| Message Is Reply       | Yes or No                                                     | Message Create, Message Update                                          |
| Message Mentions Bot   | Yes or No, whether the message pings your bot                 | Message Create, Message Update                                          |
| User Is Bot            | Yes or No. Choose No to ignore bots                           | Member Join, Member Leave, Reaction Add                                 |

The editor only offers the targets the event of your listener has. A filter on something the event doesn't have, like the message content of a Member Join, would stop the listener from ever running, so it can't be saved. Messages from bots are always ignored, which is why User Is Bot isn't offered for message events.

Message Content filters are case-sensitive. Enable `Ignore Case` so that `!help` also matches `!Help`.

Filtering on the message ID and the reaction emoji is useful for reaction roles: the flow only runs for the right emoji on the right message.

## How Filters Run

Kite checks filters before the flow starts. Events that don't match don't run any blocks, unlike a condition block, which only runs once the flow has started. If a listener only cares about some events, like messages in one channel, a filter is the better choice.

<NodeInfoExplorer type="option_event_filter" />

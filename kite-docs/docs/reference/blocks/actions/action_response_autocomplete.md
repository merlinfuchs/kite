---
sidebar_position: 6.5
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Show autocomplete choices

<EmbedFlowNode type="action_response_autocomplete" />

The `Show autocomplete choices` block shows the choices for what the user is typing in a [command argument](../options/option_command_argument.md#autocomplete) with autocomplete. It only works in the blocks after the argument's autocomplete output.

Those blocks run every time the user types in the argument, before they send the command. They can look things up, like a stored variable or an API, but can't respond in any other way, like sending a message.

## Fields

### Choices

The choices to show, each with a name the user sees and a value the command receives when it's picked. Both can use placeholders, and a choice without a value uses its name. `{{autocomplete.value}}` is what the user typed so far, so a choice named `{{autocomplete.value}}` lets them pick exactly what they typed.

### Choices From List

A list to show more choices from, like `{{var('fruits')}}` or the result of an API request. Its items can be texts or numbers, which are both the name and value of a choice, or objects with a `name` and `value`. Text that isn't a list has a choice on each line.

### Only Matching Choices

Only shows the choices that contain what the user typed, ignoring case, those that start with it first. Turn it off when the list is already filtered, e.g. by an API that searches for what was typed.

## Limits

Discord shows up to 25 choices, so only the first 25 are shown. Names longer than 100 characters are cut, and choices whose value is longer than 100 characters are left out. For integer and number arguments, choices whose value isn't a number are left out.

Every time the blocks run they use credits, like other blocks, and they have 2.5 seconds to show the choices. Without a `Show autocomplete choices` block, no choices are shown.

Autocomplete only suggests values: the user can still send the command with anything they typed, so check the value in the command if it has to be one of the choices.

<NodeInfoExplorer type="action_response_autocomplete" />

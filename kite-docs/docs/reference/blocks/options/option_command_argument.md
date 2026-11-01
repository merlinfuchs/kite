---
sidebar_position: 45
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Command Argument

<EmbedFlowNode type="option_command_argument" />

The `Command Argument` block allows you to define arguments for your slash commands. You can specify the argument name, description, and type to create interactive command interfaces.

## Fields

### Name 
Here you will enter the name for your argument.

### Description 
Here you will enter a description for your argument.

### Type
Here you can choose what kind of option you want to use.

### List of types:
> `Text`: Allows the user to insert blank text.
> 
> `Whole number`: Allows the user to insert a whole number.
>
> `True/False`: Allows the user to choose between True and False.
>
> `User`: Allows the user to choose a user.
>
> `Channel`: Allows the user to choose a channel.
>
> `Mentionable`: Allows the user to choose a role or user.
>
> `Attachment`: Allows the user to attach a file.

### Choices
Choices allow the user to select a value from a predefined set of options.

Use the plus button to add a choice and the minus button next to a choice to remove just that one. The trash button removes all choices. Discord allows up to 25 choices, turn on [autocomplete](#autocomplete) for more.

### Autocomplete
Text, whole number and number arguments can suggest choices while the user types, instead of showing a fixed list.

Without blocks connected to the argument's `Autocomplete` output, the choices of the argument that contain what was typed are shown, up to 500 of them. You can add them one by one, or paste them with one choice per line, as `name = value` or just a name. This doesn't use credits.

To show other choices, like ones from a stored variable or an API, connect blocks to the `Autocomplete` output on the right of the argument and end them with a [Show autocomplete choices](../actions/action_response_autocomplete.md) block. They run every time the user types in the argument, and `{{autocomplete.value}}` is what was typed so far. Other arguments the user already filled in are available with `{{arg('ARGNAME')}}`, where users, roles and channels are their IDs.

Autocomplete only suggests values: the user can still send the command with anything they typed, so check the value in the command if it has to be one of the choices.

## Order

Arguments show up in Discord in the order set in the [Command](../entries/entry_command#argument-order) block. Required arguments always come before optional ones.

## Output

To use the result of your option later in your flow, you will use the variable
`{{arg('ARGNAME')}}`.

:::tip

`ARGNAME` should be replaced with the name of your chosen argument.

:::

<NodeInfoExplorer type="option_command_argument" />

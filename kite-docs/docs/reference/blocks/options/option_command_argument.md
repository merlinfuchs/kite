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
> ` Mentionable`: Allows the user to choose a role or user.
>
> `Attatchment`: Allows the user to attatch a file.

### Choices
Choices allow the user to select a value from a predefined set of options.

## Output

To use the result of your option later in your flow you will use the variable
`{{arg('ARGNAME')}}`.

:::tip

`ARGNAME` should be replaced with the name of your chosen argument.

:::

<NodeInfoExplorer type="option_command_argument" />

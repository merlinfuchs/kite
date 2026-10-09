---
sidebar_position: 7
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Show Modal

<EmbedFlowNode type="suspend_response_modal" />

Instead of creating a message response you can also show a modal to the user to ask for further information. Modals can have a number of inputs which you can then access using the `input(...)` variables once the modal has been submitted.

## Components

A modal can hold up to 5 components, and at least one of them has to be an input.

- **Input**: a label, an optional description below it, and one input. Every input has an identifier that you read its answer with, e.g. `{{input('color')}}`.
- **Text**: markdown text shown between the inputs, e.g. to explain the form.

These inputs are available:

| Input               | Answer                                                             |
| ------------------- | ------------------------------------------------------------------ |
| Text Input          | The entered text. Can be a single line or a paragraph.             |
| Select Menu         | The values of the picked options.                                  |
| User Select         | The IDs of the picked members.                                     |
| Role Select         | The IDs of the picked roles.                                       |
| User or Role Select | The IDs of the picked members or roles.                            |
| Channel Select      | The IDs of the picked channels. You can limit it to channel types. |
| Radio Group         | The value of the picked option. Needs 2 to 10 options.             |
| Checkbox Group      | The values of the picked options. Needs 1 to 10 options.           |
| Checkbox            | `true` if it was checked, otherwise `false`.                       |

An option without a value returns its label. Select menus let one option be picked unless you raise Max Picks, and checkbox groups allow all of their options. When several are picked, `input(...)` returns the first one, like `interaction.value` of a select menu in a message. `inputs(...)` returns all of them as a list:

| Example                            | Result if `red` and `blue` were picked |
| ---------------------------------- | -------------------------------------- |
| `{{'red' in inputs('colors')}}`    | `true`                                 |
| `{{len(inputs('colors'))}}`        | `2`                                    |
| `{{join(inputs('colors'), ', ')}}` | `red, blue`                            |

`inputs(...)` works for every input. For a text input, radio group or checkbox it returns a list with the one value.

Inputs that aren't required can be left empty, in which case `input(...)` returns an empty text and `inputs(...)` an empty list.

:::info
File uploads in modals aren't supported yet.
:::

Responding with a modal starts a sub-flow which is suspended until the user submits the modal. See [Sub-Flows](/reference/sub-flows) for more information on how modals work.

:::tip
The answers stay available as `input(...)` and `inputs(...)` even after later buttons or modals. Command arguments keep working with `arg(...)` below the modal too.
:::

<NodeInfoExplorer type="suspend_response_modal" />

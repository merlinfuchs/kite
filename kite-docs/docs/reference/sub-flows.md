---
sidebar_position: 6
---

# Sub-flows

Some flow blocks can be used to create sub-flows. These blocks act as a boundary between the main flow and the sub-flow and are highlighted in pink in the flow editor.

## Placeholders in Sub-Flows

A sub-flow runs with the interaction that resumed it. Inside the sub-flow of a button, `user` and `interaction` refer to whoever clicked the button, which on a public message isn't necessarily the user who ran the command. Node results and temporary variables from before the sub-flow remain available.

To access the interaction or event from before the sub-flow, use `origin` and `previous`:

- `origin` is the interaction or event that started the flow, e.g. `{{origin.user.mention}}` is the user who ran the command.
- `previous` is the interaction that led to the current sub-flow. It's only different from `origin` when sub-flows are nested, e.g. a button inside the sub-flow of a modal.

`arg()` and `input()` keep working in sub-flows. `arg()` returns the argument of the command that started the flow, and `input()` returns the value from the modal it was submitted in, up to three modals back.

## Modals

![Modal Node](./img/example-node-modal.png)

Modals are a special type of sub-flow. They are used to create interactive experiences. When a modal is opened, the flow execution is suspended and the modal is displayed to the user. The flow execution is resumed when the modal is submitted. You can then access the results of the modal in the main flow using the `interaction.components` placeholder.

Modals can be used to create interactive experiences like forms, quizzes, etc.

## Interactive Messages

![Interactive Message Node](./img/example-node-message-buttons.png)

When adding buttons or select menus to a message, the message becomes interactive. When an interactive message is sent and a user interacts with it, the execution resumes from the corresponding button or select menu in the flow. Just attach the blocks you want to each button or select menu.

In the sub-flow of a select menu the value of the picked option is available as `{{interaction.value}}`.

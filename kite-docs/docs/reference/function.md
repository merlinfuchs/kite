---
sidebar_position: 3.5
---

# Functions

Functions let you build blocks you need in several places once, and run them from any command, event listener, message template or other function. When you change a function, every flow that calls it uses the new version right away, without saving or deploying them again.

You create functions on the **Functions** page of your app. Each function has its own flow, which starts with the [Function](./blocks/entries/entry_function.md) block.

## Calling a function

Add the [Call function](./blocks/actions/action_function_call.md) block to a flow and pick the function. The function runs right there, and the blocks after the Call function block run once it has finished.

A function runs with the interaction or event of the flow that called it. That means `{{user}}`, `{{guild}}` and `{{channel}}` are the same as in the caller, and a function called from a command can respond to it, e.g. with the `Create response message` block.

## Parameters

Parameters are the values a function receives. Add them in the settings of the Function block, and read them in the function with `{{arg('name')}}`, just like command arguments.

The Call function block shows an input for each parameter of the function you picked. Inputs can contain placeholders, so you can pass anything from the calling flow, like `{{arg('target').id}}` in a command or `{{result('abc').content}}`.

## Returning a value

The [Return](./blocks/controls/control_function_return.md) block ends the function and sends a value back. In the calling flow, that value is the result of the Call function block, e.g. `{{result('abc')}}`, or the temporary variable you name in the block's settings.

Nothing after a Return block runs in the function. A function without a Return block runs all its blocks and returns nothing.

## Stopping the calling flow

Turn on **Stop Calling Flow** in a Return block to stop the flow that called the function as well. Nothing after the Call function block runs then. This makes checks reusable, like a permission check:

1. Create a function, e.g. `Staff Only`, with a `User Condition` on `{{user}}`.
2. On its **Has Role** branch, put nothing, so the function just ends and the command continues.
3. On its **Else** branch, add a `Create response message` block saying "You don't have permission", followed by a Return block with **Stop Calling Flow** turned on.

Every command that starts with a Call function block for `Staff Only` now only runs for staff, and the role and message are in one place.

## What stays in the function

Results of blocks and temporary variables are separate for every call. A function can't see the results of the blocks in the flow that called it, so pass what it needs as parameters. Stored variables are shared, as they are everywhere in your app.

If a block in the function fails, the Call function block fails with the same error, so you can handle it with a `Handle Errors` block around the Call function block.

## Deleting a function

The Functions page shows which commands, event listeners, message templates and other functions call each function. When you delete a function that's still used, Kite lists them first, as their Call function blocks fail until you pick another function.

## Duplicating and sharing

Click on the `...` next to a function to duplicate it with a new name, or to export it. Commands, event listeners and message templates that call functions export them along with them, see [Templates and Sharing](./templates.md#sharing-commands-event-listeners-message-templates-and-functions).

## Limits

- Functions can call other functions, and themselves, up to 10 levels deep.
- The blocks of a function cost credits like everywhere else and count towards the limits of the flow that called it. The Call function block itself is free.
- Functions run within the flow that calls them, so they can't wait for something that happens later: the `Show Modal` block, buttons and select menus with blocks attached, and `Wait` blocks longer than 5 seconds don't work in functions. Put them in the calling flow instead, or send a message template, whose buttons have their own blocks.
- An app can have up to 100 functions.

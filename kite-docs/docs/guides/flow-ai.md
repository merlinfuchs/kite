---
sidebar_position: 4
---

# Building Flows with AI

The flow AI builds and changes flows for you. Describe what the flow should do, and it adds the blocks and connects them. You can also ask it how something works.

Open a command or event listener and click on `Ask AI` at the top right of the editor. For the flow of a button or select menu, the button is at the top left.

## Asking for Changes

Write what you want, like "Ban the user from the argument and log why" or "Add a cooldown of 10 seconds". The AI changes the flow right away and highlights the blocks it changed.

- **Nothing is saved yet.** Look at the changes, then click on `Save Changes`. For commands, click on `Deploy Changes` afterwards if the name, description or arguments changed.
- **Undo** reverts everything the AI changed for one message.
- **Select blocks** before asking to ask about them, like "Why does this block fail?" or "Make this message red".
- When the AI needs something only you know, like a channel or role, it shows fields to fill in instead of guessing. Click on `Send` when you're done.
- When the AI suggests a change without making it, click on `Build this` to put the request into the chat, then send it.

Kite checks the flow after every change. If the AI's changes cause problems, like a missing setting, it tries to fix them itself.

## What It Can't Do

The AI only changes the flow you have open. It can't create other commands, event listeners, message templates or stored variables. When your request needs them, it tells you what to create, and you can then open each one and ask it to build that part.

The chat isn't saved. It's gone when you leave or reload the page, or click on `New chat`. Start a new chat when you ask for something unrelated.

## Prompts

Every message that changes the flow uses one prompt. Questions, answers without changes and the AI fixing its own changes don't use prompts, but you can ask at most three times as many questions as you have prompts. The AI doesn't use any [credits](../reference/credit-system.md).

The free plan includes 5 prompts per app and month, and [Premium](../reference/premium.md) plans include more. You can see how many are left below the chat. They reset on the first day of each month.

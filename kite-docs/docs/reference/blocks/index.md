---
sidebar_position: 1
---

# Blocks

Each flow in Kite consists of a number of blocks that are connected. The blocks are executed from top to bottom. When multiple blocks are connected to the same parent node the order of execution is not defined.

The output of previously executed blocks is available in all subsequent blocks and can be accessed using the `result(...)` variables. In case there is an error in a block the execution will stop and no further blocks will be executed.

## Entry Blocks

- [Command](./entries/entry_command.md) - Entry point for slash commands
- [Listen for Event](./entries/entry_event.md) - Entry point for event-triggered flows
- [Button](./entries/entry_component_button.md) - Entry point for button component interactions

## Response Blocks

- [Create response message](./actions/action_response_create.md) - Respond to commands or interactions
- [Edit response message](./actions/action_response_edit.md) - Edit previously created responses
- [Delete response message](./actions/action_response_delete.md) - Delete response messages
- [Show Modal](./actions/suspend_response_modal.md) - Show modal dialogs to users
- [Defer response](./actions/action_response_defer.md) - Defer response for longer processing

## Message Blocks

- [Create channel message](./actions/action_message_create.md) - Send messages to channels
- [Edit channel message](./actions/action_message_edit.md) - Edit channel messages
- [Delete channel message](./actions/action_message_delete.md) - Delete channel messages
- [Get channel message](./actions/action_message_get.md) - Retrieve channel messages
- [Send direct message](./actions/action_private_message_create.md) - Send private messages
- [Create message reaction](./actions/action_message_reaction_create.md) - Add reactions to messages
- [Delete message reaction](./actions/action_message_reaction_delete.md) - Remove reactions from messages

## User & Member Blocks

- [Get user](./actions/action_user_get.md) - Retrieve user information
- [Get member](./actions/action_member_get.md) - Get server member information
- [Ban member](./actions/action_member_ban.md) - Ban members from servers
- [Unban member](./actions/action_member_unban.md) - Unban members from servers
- [Kick member](./actions/action_member_kick.md) - Kick members from servers
- [Timeout member](./actions/action_member_timeout.md) - Timeout members
- [Edit member nickname](./actions/action_member_edit.md) - Edit member nicknames
- [Add member role](./actions/action_member_role_add.md) - Add roles to members
- [Remove member role](./actions/action_member_role_remove.md) - Remove roles from members

## Server & Channel Blocks

- [Get role](./actions/action_role_get.md) - Retrieve role information
- [Get server](./actions/action_guild_get.md) - Get server information
- [Get channel](./actions/action_channel_get.md) - Retrieve channel information

## Variable Blocks

- [Set stored variable](./actions/action_variable_set.md) - Set persistent variables
- [Get stored variable](./actions/action_variable_get.md) - Retrieve stored variables
- [Delete stored variable](./actions/action_variable_delete.md) - Remove stored variables

## AI Blocks

- [Ask AI](./actions/action_ai_chat_completion.md) - Interact with AI models
- [Search the Web](./actions/action_ai_web_search.md) - Search the internet with AI

## Utility Blocks

- [Calculate Value](./actions/action_expression_evaluate.md) - Evaluate expressions and calculations
- [Generate Random Number](./actions/action_random_generate.md) - Generate random numbers
- [Send API Request](./actions/action_http_request.md) - Make HTTP requests
- [Log Message](./actions/action_log.md) - Log messages for debugging

## Roblox Blocks

- [Get Roblox User](./actions/action_roblox_user_get.md) - Retrieve Roblox user information

## Control Flow Blocks

- [Comparison Condition](./controls/control_condition_compare.md) - Create conditional logic
- [User Condition](./controls/control_condition_user.md) - User-based conditions
- [Channel Condition](./controls/control_condition_channel.md) - Channel-based conditions
- [Role Condition](./controls/control_condition_role.md) - Role-based conditions
- [Run a loop](./controls/control_loop.md) - Execute actions multiple times
- [Exit loop](./controls/control_loop_exit.md) - Exit loops early
- [Wait](./controls/control_sleep.md) - Pause flow execution

## Command Options

- [Command Argument](./options/option_command_argument.md) - Define command arguments
- [Command Permissions](./options/option_command_permissions.md) - Set command permissions
- [Command Contexts](./options/option_command_contexts.md) - Define command availability

## Event Options

- [Event Filter](./options/option_event_filter.md) - Filter events based on properties

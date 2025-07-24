---
sidebar_position: 7
---

import EmbedFlowNode from "../../src/components/EmbedFlowNode";

# Blocks

Each flow in Kite consists of a number of blocks that are connected. The blocks are executed from top to bottom. When multiple blocks are connected to the same parent node the order of execution is not defined.

The output of previously executed blocks is available in all subsequent blocks and can be accessed using the `result(...)` variables. In case there is an error in a block the execution will stop and no further blocks will be executed.

## Entry Blocks

### Command {#entry_command}

The `Command` block is the entry point for slash commands. This is where your command flow begins when a user invokes your slash command.

You can configure the command name and description directly in this block. The command will be automatically registered with Discord when you deploy your app.

<EmbedFlowNode type="entry_command" />

### Listen for Event {#entry_event}

The `Listen for Event` block is the entry point for event-triggered flows. This block listens for specific Discord events and triggers your flow when they occur.

You can configure which event type to listen for and add filters to control when the flow should be triggered.

<EmbedFlowNode type="entry_event" />

### Button {#entry_component_button}

The `Button` block is the entry point for button component interactions. This block gets triggered when a user clicks a button that was created by your bot.

This is typically used in conjunction with interactive components in your messages.

<EmbedFlowNode type="entry_component_button" />

## Response Blocks

### Create response message {#action_response_create}

The `Create response message` block is used to respond to a command or component interaction. Usually each command or component interaction must create a message response. The only exception is when you show a modal instead.

You can either configure the message in the block directly or use a message template instead. In both cases you can add embeds and interactive components to the message. The only case where it's better to use a message template is when you want to use the same response in multiple places.

If the message contains interactive components, the flow will be suspended until the user interacts with the message. See [Sub-Flows](/reference/sub-flows) for more information on how interactive components work.

<EmbedFlowNode type="action_response_create" />

### Edit response message {#action_response_edit}

The `Edit response message` block is used to edit a previously created response message. Right now you can only edit the original (first) response message.

You can either configure the message in the block directly or use a message template instead. In both cases you can add embeds and interactive components to the message. The only case where it's better to use a message template is when you want to use the same response in multiple places.

If the message contains interactive components, the flow will be suspended until the user interacts with the message. See [Sub-Flows](/reference/sub-flows) for more information on how interactive components work.

<EmbedFlowNode type="action_response_edit" />

### Delete response message {#action_response_delete}

The `Delete response message` block is used to delete a previously created response message. Right now you can only delete the original (first) response message.

<EmbedFlowNode type="action_response_delete" />

### Show Modal {#suspend_response_modal}

Instead of creating a message response you can also show a modal to the user to ask for further information. Modals can have a number of inputs which you can then access using the `input(...)` variables once the modal has been submitted.

Responding with a modal starts a sub-flow which is suspended until the user submits the modal. See [Sub-Flows](/reference/sub-flows) for more information on how modals work.

<EmbedFlowNode type="suspend_response_modal" />

### Defer response {#action_response_defer}

If you know that your flow takes longer than 3 seconds before it can respond, you can use the `Defer response` block to let Discord know that the response is taking longer. After that you have up to 15 minutes to respond.

This block is usually not necessary and can be omitted. Kite is smart enough to detect when your flow is taking too long and will defer the response automatically.

<EmbedFlowNode type="action_response_defer" />

## Message Blocks

### Create channel message {#action_message_create}

The `Create channel message` block is used to send a message to a specific channel.

You can either configure the message in the block directly or use a message template instead. In both cases you can add embeds and interactive components to the message. The only case where it's better to use a message template is when you want to use the same response in multiple places.

If the message contains interactive components, the flow will be suspended until the user interacts with the message. See [Sub-Flows](/reference/sub-flows) for more information on how interactive components work.

<EmbedFlowNode type="action_message_create" />

### Edit channel message {#action_message_edit}

The `Edit channel message` block is used to edit a message in a specific channel.

You can either configure the message in the block directly or use a message template instead. In both cases you can add embeds and interactive components to the message. The only case where it's better to use a message template is when you want to use the same response in multiple places.

If the message contains interactive components, the flow will be suspended until the user interacts with the message. See [Sub-Flows](/reference/sub-flows) for more information on how interactive components work.

<EmbedFlowNode type="action_message_edit" />

### Delete channel message {#action_message_delete}

The `Delete channel message` block is used to delete a message from a specific channel.

<EmbedFlowNode type="action_message_delete" />

### Get channel message {#action_message_get}

The `Get channel message` block is used to get a message from a specific channel.

<EmbedFlowNode type="action_message_get" />

### Send direct message {#action_private_message_create}

The `Send direct message` block is used to send a message to a specific user.

You can either configure the message in the block directly or use a message template instead. In both cases you can add embeds and interactive components to the message. The only case where it's better to use a message template is when you want to use the same response in multiple places.

<EmbedFlowNode type="action_private_message_create" />

### Create message reaction {#action_message_reaction_create}

The `Create message reaction` block is used to add a reaction to a specific message.

The emoji can be a standard emoji or a custom emoji. The emoji picker only shows custom emojis that are uploaded to the app by going to the `Emojis` tab in the Kite dashboard.

<EmbedFlowNode type="action_message_reaction_create" />

### Delete message reaction {#action_message_reaction_delete}

The `Delete message reaction` block is used to remove a reaction from a specific message.

The emoji can be a standard emoji or a custom emoji. The emoji picker only shows custom emojis that are uploaded to the app by going to the `Emojis` tab in the Kite dashboard.

<EmbedFlowNode type="action_message_reaction_delete" />

## User Blocks

### Get user {#action_user_get}

The `Get user` block is used to get a user by their ID.

<EmbedFlowNode type="action_user_get" />

## Member Blocks

### Get member {#action_member_get}

The `Get member` block is used to get a member by their ID from a specific server.

<EmbedFlowNode type="action_member_get" />

### Ban member {#action_member_ban}

The `Ban member` block is used to ban a member from a server.

<EmbedFlowNode type="action_member_ban" />

### Unban member {#action_member_unban}

The `Unban member` block is used to unban a member from a server.

<EmbedFlowNode type="action_member_unban" />

### Kick member {#action_member_kick}

The `Kick member` block is used to kick a member from a server.

<EmbedFlowNode type="action_member_kick" />

### Timeout member {#action_member_timeout}

The `Timeout member` block is used to timeout a member from a server.

<EmbedFlowNode type="action_member_timeout" />

### Edit member nickname {#action_member_edit}

The `Edit member nickname` block is used to edit the nickname of a member.

Right now only the nickname can be edited. In the future we will add more options.

<EmbedFlowNode type="action_member_edit" />

### Add member role {#action_member_role_add}

The `Add member role` block is used to add a role to a member.

<EmbedFlowNode type="action_member_role_add" />

### Remove member role {#action_member_role_remove}

The `Remove member role` block is used to remove a role from a member.

<EmbedFlowNode type="action_member_role_remove" />

## Role Blocks

### Get role {#action_role_get}

The `Get role` block is used to get a role by its ID from a specific server.

<EmbedFlowNode type="action_role_get" />

## Server Blocks

### Get server {#action_guild_get}

The `Get server` block is used to get a server (guild) by its ID.

<EmbedFlowNode type="action_guild_get" />

## Channel Blocks

### Get channel {#action_channel_get}

The `Get channel` block is used to get a channel by its ID.

<EmbedFlowNode type="action_channel_get" />

## Stored Variable Blocks

### Set stored variable {#action_variable_set}

The `Set stored variable` block is used to set the value of a stored variable. Stored variables persist across flow executions and can be used to maintain state or store data for later use.

You can choose from different scopes (user, server, global) and operations (set, increment, decrement, append, etc.).

<EmbedFlowNode type="action_variable_set" />

### Get stored variable {#action_variable_get}

The `Get stored variable` block is used to retrieve the value of a stored variable. The value can then be used in subsequent blocks using the `result(...)` variables.

<EmbedFlowNode type="action_variable_get" />

### Delete stored variable {#action_variable_delete}

The `Delete stored variable` block is used to remove a stored variable from the specified scope.

<EmbedFlowNode type="action_variable_delete" />

## AI Blocks

### Ask AI {#action_ai_chat_completion}

The `Ask AI` block allows you to interact with artificial intelligence models. You can ask questions, get responses to prompts, or have the AI perform various text-based tasks.

This block supports different AI models with varying capabilities and costs. The response can be used in subsequent blocks.

<EmbedFlowNode type="action_ai_chat_completion" />

### Search the Web {#action_ai_web_search}

The `Search the Web` block allows you to search the internet for the latest information using AI. This is useful for getting current information that might not be available in the AI's training data.

This block uses AI to process and summarize web search results, providing you with relevant and up-to-date information.

<EmbedFlowNode type="action_ai_web_search" />

## Utility Blocks

### Calculate Value {#action_expression_evaluate}

The `Calculate Value` block allows you to evaluate mathematical expressions or perform logical operations. You can use this to perform calculations, string manipulations, or other data processing tasks.

See [Expressions](/reference/expressions) for more information on what you can do with expressions.

The result can be stored in a temporary variable and used in subsequent blocks.

<EmbedFlowNode type="action_expression_evaluate" />

### Generate Random Number {#action_random_generate}

The `Generate Random Number` block generates a random number within a specified range. This is useful for creating games, random selections, or adding unpredictability to your flows.

<EmbedFlowNode type="action_random_generate" />

### Send API Request {#action_http_request}

The `Send API Request` block allows you to make HTTP requests to external APIs or web services. You can send GET, POST, PUT, DELETE, and other HTTP methods with custom headers and body data.

This is useful for integrating with external services, fetching data from APIs, or triggering webhooks.

<EmbedFlowNode type="action_http_request" />

### Log Message {#action_log}

The `Log Message` block allows you to log text messages that are only visible in the application logs. This is useful for debugging, monitoring, or keeping track of flow execution.

You can choose different log levels (info, warning, error) to categorize your log messages.

<EmbedFlowNode type="action_log" />

## Roblox Blocks

### Get Roblox User {#action_roblox_user_get}

The `Get Roblox User` block allows you to retrieve information about a Roblox user by their ID or username. This is useful for creating integrations with Roblox or building features that work with Roblox user data.

You can look up users by their Roblox ID or username, and the block will return information about the user.

<EmbedFlowNode type="action_roblox_user_get" />

## Control Flow Blocks

### Comparison Condition {#control_condition_compare}

The `Comparison Condition` block allows you to create conditional logic based on comparing two values. You can set up multiple comparison conditions that will execute different actions based on the results.

This block creates a branching structure where different paths can be taken depending on the comparison results.

<EmbedFlowNode type="control_condition_compare" />

#### Match Condition {#control_condition_item_compare}

The `Match Condition` block is used within a comparison condition to define specific comparison criteria. It runs actions if the two values being compared meet the specified conditions.

<EmbedFlowNode type="control_condition_item_compare" />

### User Condition {#control_condition_user}

The `User Condition` block allows you to create conditional logic based on user properties or characteristics. You can set up multiple user-based conditions that will execute different actions.

This is useful for creating user-specific behavior or permission-based flows.

<EmbedFlowNode type="control_condition_user" />

#### Match User {#control_condition_item_user}

The `Match User` block is used within a user condition to define specific user matching criteria. It runs actions if the user meets the specified conditions.

<EmbedFlowNode type="control_condition_item_user" />

### Channel Condition {#control_condition_channel}

The `Channel Condition` block allows you to create conditional logic based on channel properties or characteristics. You can set up multiple channel-based conditions that will execute different actions.

This is useful for creating channel-specific behavior or restricting certain actions to specific channels.

<EmbedFlowNode type="control_condition_channel" />

#### Match Channel {#control_condition_item_channel}

The `Match Channel` block is used within a channel condition to define specific channel matching criteria. It runs actions if the channel meets the specified conditions.

<EmbedFlowNode type="control_condition_item_channel" />

### Role Condition {#control_condition_role}

The `Role Condition` block allows you to create conditional logic based on role properties or characteristics. You can set up multiple role-based conditions that will execute different actions.

This is useful for creating role-based permissions or role-specific behavior.

<EmbedFlowNode type="control_condition_role" />

#### Match Role {#control_condition_item_role}

The `Match Role` block is used within a role condition to define specific role matching criteria. It runs actions if the role meets the specified conditions.

<EmbedFlowNode type="control_condition_item_role" />

#### Else {#control_condition_item_else}

The `Else` block is used within conditional structures to define actions that should run when no other conditions are met. This provides a fallback path for your conditional logic.

<EmbedFlowNode type="control_condition_item_else" />

### Run a loop {#control_loop}

The `Run a loop` block allows you to execute a set of actions multiple times. You can specify how many times the loop should run, and the actions within the loop will be repeated accordingly.

This is useful for processing lists, performing repetitive tasks, or creating iterative workflows.

<EmbedFlowNode type="control_loop" />

#### Each loop iteration {#control_loop_each}

The `Each loop iteration` block defines the actions that should be executed for each iteration of the loop. This block is automatically created when you add a loop block.

<EmbedFlowNode type="control_loop_each" />

#### After loop {#control_loop_end}

The `After loop` block defines actions that should be executed after the loop has finished running. This block is automatically created when you add a loop block.

<EmbedFlowNode type="control_loop_end" />

### Exit loop {#control_loop_exit}

The `Exit loop` block allows you to exit out of a loop before it has completed all iterations. This is useful for creating early termination conditions or breaking out of loops when certain criteria are met.

<EmbedFlowNode type="control_loop_exit" />

### Wait {#control_sleep}

The `Wait` block pauses the flow execution for a specified amount of time. This is useful for creating delays, rate limiting, or scheduling actions to occur after a certain period.

<EmbedFlowNode type="control_sleep" />

## Command Options

### Command Argument {#option_command_argument}

The `Command Argument` block defines arguments that can be passed to your slash command. You can specify the argument name, description, type, and whether it's required.

This block allows users to provide input to your command when they invoke it.

<EmbedFlowNode type="option_command_argument" />

### Command Permissions {#option_command_permissions}

The `Command Permissions` block allows you to restrict your command to users with specific permissions. You can specify which Discord permissions are required to use the command.

This is useful for creating admin-only commands or commands that require specific server permissions.

<EmbedFlowNode type="option_command_permissions" />

### Command Contexts {#option_command_contexts}

The `Command Contexts` block allows you to define where your command should be available. You can specify whether the command should be available in servers, DMs, or both, and which integrations should have access to it.

By default, commands are available everywhere, but you can restrict them to specific contexts.

<EmbedFlowNode type="option_command_contexts" />

## Event Options

### Event Filter {#option_event_filter}

The `Event Filter` block allows you to filter events based on their properties. You can specify conditions that must be met for the event to trigger your flow.

This is useful for creating more specific event handlers that only respond to certain types of events or events with specific characteristics.

<EmbedFlowNode type="option_event_filter" />

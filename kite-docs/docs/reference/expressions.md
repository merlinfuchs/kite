---
sidebar_position: 5
---

# Expressions

Kite supports performing calculations and transformations on data using expressions. Expressions are available in the `Evaluate Expression` block (**which also supports multi-line input**) and in every placeholder surrounded by `{{` and `}}`.

## Expression Syntax

Expressions are powered by the [Expr](https://expr-lang.org) language, you can learn more about the features and syntax [here](https://expr-lang.org/docs/language-definition).

## Available Variables

Kite provides the following variables in the expression environment:

```yaml
user:
  id: string
  username: string
  discriminator: string
  display_name: string
  avatar_url: string
  banner_url: string
  mention: string
  is_bot: bool
  created_at: int # When the account was created, as a Unix timestamp in seconds
  role_ids?: []string
  nick?: string
  # Only filled inside a server
  joined_at: int # When the user joined the server, as a Unix timestamp in seconds
  top_role: # The highest role of the user, or @everyone without roles
    id: string
    name: string
    mention: string
  role_mentions: string # Mentions of all roles of the user, highest first
  role_names: string # Names of all roles of the user, highest first, separated by commas
  role_count: int
  color: string # Color of the highest role that has one, like #5865f2
  is_booster: bool
  boosting_since: int # Unix timestamp in seconds, 0 if the user isn't boosting
  is_timed_out: bool
  timeout_until: int # Unix timestamp in seconds, 0 if the user isn't timed out
  is_owner: bool # Whether the user owns the server
  is_admin: bool # Whether the user owns the server or has the Administrator permission
  permissions: []string # Lowercased permission names, like ban_members or manage_messages

message?: # For message events
  id: string
  content: string

channel:
  id: string
  name: string
  mention: string
  type: string # text, voice, category, announcement, thread, stage, forum, media, dm, group_dm, or unknown for newer types
  category_id: string # Empty outside of a category, for a thread the category of its channel
  category_name: string

guild?: # For events and interactions inside a server, elsewhere guild is empty and so are its fields
  id: string # The id of the server
  name: string
  icon_url: string # Empty if the server has no icon
  member_count: int
  boost_count: int
  owner_id: string # The id of the user that owns the server
  boost_level: int # 0 to 3
  created_at: int # When the server was created, as a Unix timestamp in seconds
  banner_url: string # Empty if the server has no banner
  description: string
  vanity_url: string # The custom invite link, empty if the server has none
  role_count: int # Without @everyone
  channel_count: int # Without categories and threads
  emoji_count: int
  rules_channel: # Empty if the server has no rules channel
    id: string
    name: string
    mention: string
  system_channel: # Where Discord posts join and boost messages, empty if turned off
    id: string
    name: string
    mention: string

interaction?: # For commands and interactive components
  id: string
  value?: string # The value of the picked option in a select menu
  values?: []string # All picked values if the select menu allows picking more than one option
  components?: # For modal submissions, by input identifier
    identifier:
      value: string # What input('identifier') returns
      values: []string # What inputs('identifier') returns

app:
  user: # Access the underlying user of the app
    id: string
    mention: string
```

There are also a few special variables for accessing dynamic variables:

```py
arg('name') # Access value of a command argument
input('identifier') # Access value of a modal input, the first one if several options were picked
inputs('identifier') # Access all picked options of a modal input as a list
result('id') # Access the result of a previous block
```

`top_role`, `role_names`, `color`, `is_owner`, `is_admin` and `permissions` need the roles of the server. They work for `user`, members from command arguments, and `origin.user` and `previous.user` in sub-flows. A member from `result('id')`, like the result of a Get Member block, or from a variable doesn't know its server, so using them there fails the flow instead of returning a wrong answer. They also fail in a server your app isn't in, where only user-installed commands run.

`guild.member_count` follows members joining and leaving while your app has a Member Join or Member Leave event listener. Without one, it's the count from when your app last connected to Discord.

In [sub-flows](/reference/sub-flows), `origin` and `previous` give access to the interaction or event from before the sub-flow, e.g. `origin.user.id`.

## Examples

When using the `Evaluate Expression` block, you must omit the `{{` and `}}` from the expression.

### Get Command Argument

This will return the value of the `myarg` argument passed to the command.

```python
{{ arg('myarg') }}
```

### Get User Display Name

This will return the display name of the user who clicked the button or triggered the command or event.

```python
{{ user.display_name }}
```

### Get Message Content

This will return the content of the message that was sent.

```python
{{ message.content }}
```

### Check Reaction Emoji

In a Reaction Add or Reaction Remove event listener, this will return true if the reaction was a 👍. For a custom emoji, compare `emoji.id` instead.

```python
{{ emoji.name == "👍" }}
```

### Check if User Has Role

This will return true if the user has the role with the ID `123`.

```python
{{ "123" in user.role_ids }}
```

### Get Selected Option

This will return the value of the option that the user picked in a select menu.

```python
{{ interaction.value }}
```

If the select menu allows picking more than one option, this will return true if the user picked the option with the value `option-a`.

```python
{{ "option-a" in interaction.values }}
```

### Welcome a New Member

This will greet a member with the name of the server and their member number.

```python
Welcome to {{ guild.name }}, {{ user.mention }}! You're member #{{ guild.member_count }}.
```

### Check Picked Modal Options

If a modal input with the identifier `colors` allows picking more than one option, `inputs('colors')` returns all picked values as a list. This will return true if the user picked `red`.

```python
{{ 'red' in inputs('colors') }}
```

This will return how many options were picked.

```python
{{ len(inputs('colors')) }}
```

And this will list them separated by commas, e.g. `red, blue`.

```python
{{ join(inputs('colors'), ', ') }}
```

### Check User Creation Date

This will show when the user's account was created, as a Discord timestamp.

```python
<t:{{ user.created_at }}:f>
```

This will return true if the account is less than 7 days old, which is useful for anti-raid checks. `user.joined_at` works the same way for how long someone has been in the server.

```python
{{ user.created_at > now().Unix() - 7 * 86400 }}
```

### Only Run in a Category

This will return true if the command was used in a channel of the category named `Tickets`. Use `channel.category_id` instead to keep working when the category is renamed.

```python
{{ channel.category_name == "Tickets" }}
```

This will return true inside a thread.

```python
{{ channel.type == "thread" }}
```

### Check if User Owns the Server

This will return true if the user is the owner of the server.

```python
{{ user.is_owner }}
```

### Check the Permissions of a User

This will return true if the user is allowed to ban members. The names are the [permission names of Discord](https://discord.com/developers/docs/topics/permissions#permissions-bitwise-permission-flags) in lowercase.

```python
{{ "ban_members" in user.permissions }}
```

`user.permissions` are the permissions the roles of the user give them in the server. Permissions a single channel grants or denies aren't included. Administrators and the owner of the server have every permission, and `user.is_admin` is true for them.

### Point to the Rules Channel

This will mention the rules channel the server has set in its settings, and is empty if it has none.

```python
Please read {{ guild.rules_channel }} before posting.
```

### Show the Roles of a User

This will mention the highest role of the user, and then all of their roles.

```python
{{ user.top_role.mention }}
{{ user.role_mentions }}
```

### Do Some Math

This will return the result of the expression.

```python
{{ 1 + 1 }}
```

### Decode JSON Response

This will return the value of the `somefield` field in the JSON response of a HTTP request block with the id `owlspush`.

```python
{{ result('owlspush').data().somefield }}
```

### Timestamps

Discord timestamps are shown in the local timezone of whoever reads the message. Wrap a Unix timestamp in `<t:...:style>` and pick one of the styles below.

```python
<t:{{ now().Unix() }}:t> # Short time, like 5:36 PM
<t:{{ now().Unix() }}:T> # Long time, like 5:36:12 PM
<t:{{ now().Unix() }}:d> # Short date, like 01/02/2026
<t:{{ now().Unix() }}:D> # Long date, like January 2, 2026
<t:{{ now().Unix() }}:f> # Short date and time, like January 2, 2026 5:36 PM
<t:{{ now().Unix() }}:F> # Long date and time, like Friday, January 2, 2026 5:36 PM
<t:{{ now().Unix() }}:R> # Relative time, like 2 minutes ago
```

This will show a countdown that ends one hour from now.

```python
<t:{{ now().Add(duration("1h")).Unix() }}:R>
```

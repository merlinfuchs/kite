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
  role_ids?: []string
  nick?: string

message?: # For message events
  id: string
  content: string

channel:
  id: string

guild?: # For events and interactions inside a server
  id: string # The id of the server

interaction?: # For commands and interactive components
  id: string
  value?: string # The value of the picked option in a select menu
  values?: []string # All picked values if the select menu allows picking more than one option

app:
  user: # Access the underlying user of the app
    id: string
    mention: string
```

There are also a few special variables for accessing dynamic variables:

```py
arg('name') # Access value of a command argument
input('identifier') # Access value of a modal input
result('id') # Access the result of a previous block
```

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

### Check User Creation Date

This will show when the user's account was created, as a Discord timestamp.

```python
<t:{{ floor(((int(user.id) / 4194304) + 1420070400000) / 1000) }}:f>
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

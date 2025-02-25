---
sidebar_position: 5
---

# Expressions

Kite supports performing calculations and transformations on data using expressions. Expressions are available in the `Evaluate Expression` block and in every placeholder surrounded by `{{` and `}}`.

## Expression Syntax

Expressions are powered by the [Expr](https://expr-lang.org) language, you can learn more about the features and syntax [here](https://expr-lang.org/docs/language-definition).

## Available Variables

Kite provides the following variables in the expression environment:

```yaml
interaction: # For command and button interactions
  command: # For command interactions
    id: string
    args: # Arguments passed to the command
      arg1: any
      arg2: any
      arg3: any
      ...

  components: # For modal interactions, the values of the components
    component1: string
    component2: string

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
  channel:
    id: string
  guild?:
    id: string

event: # For event listeners
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
  channel?:
    id: string
  guild?:
    id: string

app:
  user: # Access the underlying user of the app
    id: string
    mention: string
```

## Examples

When using the `Evaluate Expression` block, you must omit the `{{` and `}}` from the expression.

### Get Command Argument

This will return the value of the `myargs` argument passed to the command.

```python
{{ interaction.command.args.myargs }}
```

### Get User Display Name

This will return the display name of the user who clicked the button or triggered the command or event.

```python
{{ interaction.user.display_name }}
```

### Get Message Content

This will return the content of the message that was sent.

```python
{{ event.message.content }}
```

### Check if User Has Role

This will return true if the user has the role with the ID `123`.

```python
{{ "123" in interaction.user.roles }}
```

### Do Some Math

This will return the result of the expression.

```python
{{ 1 + 1 }}
```

### Decode JSON Response

This will return the value of the `somefield` field in the JSON response of a HTTP request block.

```python
{{ fromJSON(nodes["owlspush"].result.body()).somefield }}
```

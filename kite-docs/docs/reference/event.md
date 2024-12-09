---
sidebar_position: 3
---

# Event Listener

With Event Listeners you can listen for events inside the Discord servers that your bot is in. Right now, Kite supports the following events:

- Message Create
- Message Update
- Message Delete
- Member Join
- Member Leave

## Restrictions

- You can only have 5 event listeners per app.
- Kite will ignore messages that are sent by a bot.
- Member events are only available when you enable the "Server Members Intent" in the [Discord Developer Portal](https://discord.dev).

![Example Event Flow](./img/example-event-flow.png)

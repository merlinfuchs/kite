---
sidebar_position: 5.5
---

# Credit Limits

Credit limits stop a single server or user from using up all of your app's [credits](./credit-system.md). Set them under _Credit Limits_ in your app, up to 50 per app.

A limit applies to servers or to users and resets every day or every month, at midnight UTC. Once a server or user has used that many credits, commands, event listeners and message buttons don't run for it until the period ends. Users who run a command or click a button get a private message telling them they reached the limit, and your app's logs get a warning the first time it happens.

## Defaults and specific limits

- A **default** limit applies to every server or user that has no limit of its own, for example _100 credits per user per day_.
- A **specific** limit applies to one server or user by its ID and replaces the default of the same period for it. Give your own server a higher limit, or a spamming user a limit of `0` to block them completely.
- A specific limit can also be set to **no limit**, which exempts that server or user from the default.

Daily and monthly limits are separate, so a server can have both, and is stopped by whichever it reaches first.

## Top usage

The _Top usage_ table on the same page shows the servers and users that used the most credits today or this month, with the limit that applies to each. Click _Set limit_ to give one a limit of its own.

## Good to know

- Only credits used after this feature was added count towards a limit.
- Executions that don't belong to a server or user aren't limited by it. Scheduled event listeners and webhooks have neither, and direct messages have no server.
- Limits are checked before a flow starts, so executions that run at the same time can go slightly over a limit together. Changes to limits take up to a minute to apply.
- Credit limits don't give your app more credits. When the app runs out of its monthly credits it's still stopped.

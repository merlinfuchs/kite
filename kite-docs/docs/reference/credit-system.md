---
sidebar_position: 5
---

# Credit System

Kite is free to use, but we do have a credit system in place to prevent abuse. Most actions taken by your app in flows will consume [a set amount](#cost-breakdown) of credits.

By default, your app has **10,000 credits available per month**. This is usually more than enough for most apps. You can always subscribe to one of the **Premium** plans to get more credits.

You can track your credit usage on the dashboard in the Monthly Usage section.
![Credit System](./img/example-usage.png)

Once your app has used 80% of its credits, the dashboard warns you. When it has used all of them, Kite stops it and your bot goes offline. It starts again by itself within a few minutes after the next month begins, which is the 1st at midnight UTC, or after you upgrade to a plan with more credits. An app that was stopped before last month stays stopped until you click on `Start App`.

## Cost Breakdown

Credits pay for what your flows do, not for what they look up. Most actions in flows will consume **1 credit per execution** with a few exceptions:

- **`Ask AI` block**:
  - Fast (default): 5 credits per execution
  - Balanced: 20 credits per execution
  - Smartest: 100 credits per execution
- **`Search The Web` block**:
  - Fast (default): 25 credits per execution
  - Balanced: 100 credits per execution
  - Smartest: 500 credits per execution
- **`Send API request` block**: 3 credits per execution
- **Blocks that only read**: free. That's `Get stored variable` and the blocks that get a server, channel, role, member, user or message, or list messages. Reading from other services, like `Get ER:LC server`, still costs 1 credit.

Control flow blocks, like conditions and loops, will not consume any credits, except `AI Choice` and `AI Check`, which cost 1 credit like `AI Score`.

## Tips

Since most actions in your flows consume credits, it’s important to run actions only when necessary. For example, you should usually not run actions on every message. Instead, you should use conditions to only run actions when certain conditions are met.

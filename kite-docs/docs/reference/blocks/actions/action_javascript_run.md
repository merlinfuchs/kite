---
sidebar_position: 34.1
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Run JavaScript

<EmbedFlowNode type="action_javascript_run" />

The `Run JavaScript` block runs a short piece of JavaScript and passes what it returns to later blocks. Use it for logic that would take many blocks, like going through a list, reading numbers out of a message or several calculations that build on each other. For a single calculation, the [Calculate Value](./action_expression_evaluate.md) block is enough.

The code runs as the body of a function. End it with `return` and the value, and later blocks read it with `{{result('js_abc')}}`, where `js_abc` is the ID of the block. It can return a number, text, a list or an object, like `{{result('js_abc').total}}`.

```js
const rolls = [];
for (let i = 0; i < input.dice; i++) {
  rolls.push(Math.floor(Math.random() * 6) + 1);
}
return { rolls: rolls.join(", "), total: rolls.reduce((a, b) => a + b, 0) };
```

## Inputs

Values from the flow go in through inputs. Each input has a name and a value, and the code reads it as `input.name`. An input named `dice` with the value `{{arg('dice')}}` is `input.dice` in the code above.

A value that is a single placeholder keeps its type, so `{{arg('dice')}}` stays a number and a list stays a list. Users, members, channels, roles and messages become objects with the same fields as their placeholders. An input `user` with the value `{{user}}` gives the code `input.user.display_name`, `input.user.mention` and `input.user.joined_at`.

A few fields are left out, because Kite only works them out when a placeholder uses them. Those are `role_names`, `role_mentions`, `permissions`, `is_admin`, `is_owner`, `top_role` and `color` of a user, the category of a channel, and the counts and channels of a server. Pass them as inputs of their own, like `{{user.is_admin}}`. `{{guild}}` and `{{channel}}` are the ID of the server and channel the flow runs in, so pass the fields you need, like `{{guild.name}}`.

The response of a [Send API Request](./action_http_request.md) block has `status_code`, `body` with the text of the response and `data` with the response read as JSON.

Placeholders in the code itself aren't filled in. Otherwise anyone who can put text in them, like a member writing a message, could change the code.

## What the code can use

The code can use the JavaScript language up to ES2023 and a few newer additions:

- `Object`, `Array`, `String`, `Number`, `Math`, `Date`, `JSON`, `RegExp`, `Symbol` and `BigInt`
- `Map`, `Set`, `WeakMap`, `WeakSet` and `WeakRef`
- Typed arrays like `Uint8Array`, `ArrayBuffer` and `DataView`
- `Proxy`, `Reflect`, generators and iterator helpers
- Newer methods like `toSorted`, `findLast`, `Object.groupBy` and `Set` methods like `union`
- `parseInt`, `parseFloat`, `encodeURIComponent`, `decodeURIComponent`, `atob` and `btoa`
- `performance.now()`

Regular expressions support named groups, lookbehind and Unicode properties like `\p{L}`, but not the `v` flag. A pattern that backtracks for too long stops at the time limit.

`console.log`, `console.info`, `console.warn` and `console.error` write to the app's logs, up to 50 lines and 4 KB per run.

The code can't send requests, wait, use timers or reach anything outside the block. Use the other blocks for that, before or after this one. So there's no `fetch`, `setTimeout` or `setInterval`, and no `require`, `import`, `process` or `Buffer`. `URL`, `URLSearchParams`, `TextEncoder`, `TextDecoder`, `crypto` and `structuredClone` aren't available either.

`async` code isn't supported. The block fails if the code returns a Promise, and code after an `await` or in `queueMicrotask` never runs.

There's no `Intl`, so `toLocaleString` on dates ignores the language and time zone, and dates are in UTC. Numbers get commas, like `1,234.5`. To show a date in a time zone, use `formatDate(date, format, timeZone)`. The date can be a `Date`, a timestamp in milliseconds or text like `2026-01-01T12:00:00Z`.

```js
return formatDate(Date.now(), "dddd [at] HH:mm", "Europe/Berlin");
// Sunday at 14:30
```

| Token | Example |
| --- | --- |
| `YYYY`, `YY` | 2026, 26 |
| `MMMM`, `MMM`, `MM`, `M` | January, Jan, 01, 1 |
| `DD`, `D` | 05, 5 |
| `dddd`, `ddd` | Monday, Mon |
| `HH`, `H` | 14, 14 |
| `hh`, `h`, `A`, `a` | 02, 2, PM, pm |
| `mm`, `m`, `ss`, `s` | 05, 5, 09, 9 |
| `Z` | +01:00 |

Text in `[brackets]` stays as it is. The time zone is a name like `America/New_York`, and defaults to `UTC`. To show a time to each member in their own time zone, use Discord's timestamps instead, like `<t:{{user.created_at}}:F>`.

## Limits

The code can be up to 4,000 characters and have up to 20 inputs, which together can be up to 1 MB. It can run for 25 ms and use 8 MB of memory. That's enough to work through a few hundred items, like the messages a `List channel messages` block returns. The value it returns can be up to 16 KB.

If the code throws an error or goes over a limit, the block fails with the error and the line it happened on. An [Error handler](../controls/control_error_handler.md) block can catch it.

The block costs [10 credits](../../credit-system.md#cost-breakdown) each time it runs.

<NodeInfoExplorer type="action_javascript_run" />

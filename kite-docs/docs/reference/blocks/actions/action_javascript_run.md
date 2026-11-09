---
sidebar_position: 34.1
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Run JavaScript

<EmbedFlowNode type="action_javascript_run" />

The `Run JavaScript` block runs a short piece of JavaScript and passes what it returns to later blocks. Use it for logic that would take many blocks, like going through a list, reading numbers out of a message or several calculations that build on each other. For a single calculation, the [Calculate Value](./action_expression_evaluate.md) block is enough.

The code runs as the body of an async function, so it can use `await`. End it with `return` and the value, and later blocks read it with `{{result('js_abc')}}`, where `js_abc` is the ID of the block. It can return a number, text, a list or an object, like `{{result('js_abc').total}}`.

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

The code can read from Discord and the app's integrations, use stored variables and send requests to other websites, see below. Changes in Discord, like sending a message, still need the other blocks, before or after this one. There's no `setTimeout` or `setInterval`, and no `require`, `import`, `process` or `Buffer`. `URL`, `URLSearchParams`, `TextEncoder`, `TextDecoder`, `crypto` and `structuredClone` aren't available either.

`await`, `async` functions and `queueMicrotask` work. There are no timers, so the code can't wait for time to pass. If it waits for a promise that never finishes, the block fails. An error in a `queueMicrotask` callback is logged and doesn't stop the code.

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

## Reading from Discord and integrations

The code can read from Discord, ER:LC and Cookie API with the endpoints of their APIs. Each one is a function that returns a promise, so use `await`:

```js
const [channel, messages] = await Promise.all([
  discord.getChannel({ channel_id: input.channel }),
  discord.listMessages({ channel_id: input.channel, limit: 50 }),
]);
return `${channel.name} has ${messages.length} recent messages`;
```

For Discord, the names are the endpoints of the [Discord API Request](./action_discord_api_request.md) block in camel case, so `get_guild_member` is `discord.getGuildMember`. ER:LC has `erlc.getServer`, with the parameters `Players`, `Staff`, `JoinLogs`, `Queue`, `KillLogs`, `CommandLogs`, `ModCalls`, `EmergencyCalls` and `Vehicles` written like that. Cookie API has `cookieApi.getCaptcha` with `captcha_id`, `cookieApi.getMinecraftUser` with `code`, and `cookieApi.getCurrentTime`.

The argument is an object with the endpoint's parameters by name, like `{ channel_id: "123", limit: 50 }` or `{ Players: true }`. A call with a wrong parameter fails without being sent or billed.

The result is the JSON the API returns, with its own field names. A member has `nick` and `user.global_name`, not the `display_name` of placeholders.

If a request fails, the promise is rejected with an `Error`, like `discord.getChannel: Unknown Channel (404)`. Its `status` is the HTTP status of the response, and for Discord its `code` is [Discord's error code](https://discord.com/developers/docs/topics/opcodes-and-status-codes#json-json-error-codes), like `10003` for an unknown channel. Errors from before the request was sent, like a wrong parameter, have neither. Catch the error with `try` and `catch`, or the block fails.

```js
try {
  return await discord.getGuildMember({ guild_id: input.guild, user_id: input.user });
} catch (e) {
  if (e.status === 404) return null;
  throw e;
}
```

Only reading works for now. Sending messages and other changes still need the other blocks. Reading from Discord is free, and reading from other integrations costs [1 credit](../../credit-system.md#cost-breakdown) each time. The integration has to be enabled in the app.

## Stored variables

Add the [stored variables](../../variable.md) the code uses under _Stored variables_, each with the name the code uses for it. The code then reads and changes them by that name:

```js
const total = await variables.increment({ name: "xp", scope: input.user.id, by: 5 });
const top = await variables.list({ name: "xp", limit: 3 });
return `You have ${total} XP. The top is ${top.map((v) => v.value).join(", ")}.`;
```

- `variables.get({ name, scope })` returns the value, or `null` if there's none.
- `variables.set({ name, scope, value })` stores a value, which can be text, a number, a list or an object.
- `variables.append({ name, scope, value })` adds `value` to the end of a list as one item, or to the end of text, and returns the new value. To start a list, set the variable to `[]` first. `variables.prepend` adds it at the start.
- `variables.increment({ name, scope, by })` adds `by`, 1 if it's left out, and returns the new number.
- `variables.delete({ name, scope })` deletes the value.
- `variables.list({ name, scope_prefix, sort, limit, skip })` returns values like the [List stored variable values](./action_variable_list.md) block, each with `scope`, `value` and `updated_at`.

Runs at the same time don't overwrite each other's `append`, `prepend` and `increment`, unlike a `get` followed by a `set`.

A scoped variable needs a `scope`, like a user ID, and an unscoped one ignores it. The name stays the same when the variable is renamed, and when the flow is part of a template, apps that add it get their own copy of the variable. Getting a value is free, the others cost [1 credit](../../credit-system.md#cost-breakdown) each, like the blocks.

## Sending requests to other websites

`fetch` sends a request to any URL, like the [Send API Request](./action_http_request.md) block. It takes the method, the headers and a text body. The response has `ok`, `status`, `statusText`, `headers.get()`, `text()` and `json()`.

```js
const res = await fetch("https://api.example.com/scores", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    Authorization: "Bearer {{secrets.API_KEY}}",
  },
  body: JSON.stringify({ user: input.user.id, points: 5 }),
});
if (!res.ok) throw new Error(`The API answered ${res.status}`);
return await res.json();
```

`{{secrets.NAME}}` in the URL, the headers or the body is replaced with the app's [secret](../../secrets.md) when the request is sent, so the code never sees its value. Other placeholders aren't filled in, pass values in as inputs instead.

The body has to be text, so turn objects into JSON with `JSON.stringify` and set the `Content-Type` header. Without that header, the body is sent as plain text. Like in a browser, a response with an error status like 404 doesn't throw. Check `res.ok` or `res.status` instead. Only a request that gets no response at all throws, like one to a website that doesn't exist. The response can be up to 1 MB, and it's read as text, so images and other files don't come through.

`fetch` calls `http.request({ url, method, headers, body })`, which returns `{ status, status_text, headers, body }` with the names of the headers in lowercase. Each request costs [3 credits](../../credit-system.md#cost-breakdown), like a Send API Request block. Test runs don't send them.

## Writing the code

The editor completes the calls, their parameters, the inputs and the stored variables of the block. To write the code in your own editor, like VS Code, download [kite.d.ts](pathname:///kite.d.ts), put it next to your file and add `/// <reference path="kite.d.ts" />` as the file's first line. The editor then knows the calls and what they take. Write the code inside `async function run() { ... }` there, as editors don't allow `return` and `await` outside a function, and paste the body into the block.

## Limits

The code can be up to 4,000 characters and have up to 20 inputs, which together can be up to 1 MB. It can run for 25 ms and use 8 MB of memory. That's enough to work through a few hundred items, like the messages a `List channel messages` block returns. The value it returns can be up to 16 KB. It can make 20 calls per run, like reads from Discord, changes of stored variables and requests with `fetch`, and a run can take 10 seconds in all. Waiting for them doesn't count against the 25 ms.

If the code throws an error or goes over a limit, the block fails with the error and the line it happened on. An [Error handler](../controls/control_error_handler.md) block can catch it.

The block costs [10 credits](../../credit-system.md#cost-breakdown) each time it runs.

<NodeInfoExplorer type="action_javascript_run" />

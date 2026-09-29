---
sidebar_position: 35
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Send API Request

<EmbedFlowNode type="action_http_request" />

The `Send API Request` block allows you to make HTTP requests to external APIs or web services. You can send GET, POST, PUT, DELETE, and other HTTP methods with custom headers and body data.

This is useful for integrating with external services, fetching data from APIs, or triggering webhooks.

## Body

Turn on **JSON body** to send a JSON body. The `Content-Type: application/json` header is set for you, unless you add one yourself. With it off, the request is sent without a body.

### Placeholders in JSON

Placeholders in a JSON body are filled in based on where they are:

- **Inside quotes** the value is inserted as text and escaped, so quotes or new lines in it can't break the JSON.
- **Outside quotes** the value keeps its type. Numbers stay numbers, lists become arrays, and an empty value becomes `null`.

```json
{
  "content": "Hello {{user.name}}!",
  "count": {{result('abc').data().count}}
}
```

Pasting a JSON document into an empty editor formats it automatically, and `Shift+Alt+F` tidies up the JSON at any time.

## Response

- **Fail on error status** stops the flow with an error when the server responds with a `4xx` or `5xx` status code. The error includes the start of the response body.
- **Transform Response** is an expression that post-processes the response before later blocks use it. The response is available as `response`, and the result of the expression becomes the result of the block.

| Field                  | Description                                          |
| ---------------------- | ---------------------------------------------------- |
| `response.status_code` | The status code, like `200`                          |
| `response.status`      | The status, like `200 OK`                            |
| `response.headers`     | The headers, like `response.headers["Content-Type"]` |
| `response.body()`      | The body as text                                     |
| `response.data()`      | The body parsed as JSON                              |

For example, `response.data().items[0].name` makes the result of the block the name of the first item, so later blocks can use `{{result('abc')}}` directly.

## Testing a request

You can try the request from the block's settings before deploying anything. Open **Configure Request** and press **Send** under **Test request**. This sends a real request, so be careful with requests that change something on the other end.

The result shows:

- The status code, how long the request took and the size of the response.
- **Result**, when you set a transform: what later blocks get from this block.
- **Body** and **Headers** of the response. JSON bodies are formatted for you.
- **Sent request**, the request exactly as it was sent, with all placeholders filled in.

If something goes wrong, the error says where: building the request, sending it, an error status (with **Fail on error status** on) or the transform. The response is still shown when it was received.

### Test values

A test isn't started by a command or an event, so there's nothing for placeholders to come from. Every placeholder the request uses is listed under **Test values**, where you can enter the value it should have during the test. You can also add values by hand, written like the placeholder without the braces:

| Test value        | Stands in for                                            |
| ----------------- | -------------------------------------------------------- |
| `user.id`         | `{{user.id}}`, or any other placeholder path             |
| `arg('term')`     | A command argument                                       |
| `input('reason')` | A modal input                                            |
| `var('token')`    | A temporary variable                                     |
| `result('abc')`   | The result of another block, also for `nodes.abc.result` |

Values are used as text, except JSON objects and arrays, `true`, `false`, `null` and short numbers. Long numbers like Discord IDs stay text so they aren't rounded. A placeholder without a test value fails the test before anything is sent.

Test values are only kept until you reload the page and are never saved with the flow. Tests are limited to 10 per minute and don't use credits.

<NodeInfoExplorer type="action_http_request" />

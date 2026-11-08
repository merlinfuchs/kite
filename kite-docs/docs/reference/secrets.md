---
sidebar_position: 4.5
---

# Secrets

Secrets keep API keys and other private values out of your flows. Put a key in a secret instead of pasting it into a block, and it isn't part of the flow when you export or share it.

Create secrets under _Secrets_ in your app, up to 10 per app. Use one in a [Send API Request](./blocks/actions/action_http_request.md) block's URL, headers, query parameters or body as `{{secrets.NAME}}`, for example a header `Authorization` with the value `Bearer {{secrets.OPENWEATHER_KEY}}`.

Secrets only work in API request blocks, in the value of a [Comparison Condition](./blocks/controls/control_condition_compare.md) branch and in the key of a [Verify webhook signature](./blocks/controls/control_webhook_verify.md) block. Everywhere else, like in messages or logs, `{{secrets.NAME}}` fails, so a secret can't be sent or shown by accident. If a request fails, the error only shows the server the request went to, and secret values are replaced with `[secret]`. Names are written in uppercase, like `API_KEY`: they start with a letter or underscore and can contain uppercase letters, numbers and underscores.

In a condition, use a secret to check a token another service sends, like a webhook's `Authorization` header. Put the header in the condition's base value and `{{secrets.NAME}}` in a branch with **Equal** or **Not Equal**. The secret has to be on its own in its placeholder, like `{{secrets.NAME}}` or `Bearer {{secrets.NAME}}`, and other modes don't work with secrets. With something like **Starts With**, or an expression that uses a part of the secret, whoever sends the header could guess the secret a character at a time.

Once saved, a value can't be read back, not even by you. To change it, enter a new one. Everyone who can edit your app can use its secrets in their blocks, so this protects keys from people you share flows with, not from your collaborators.

When you import a flow that uses secrets your app doesn't have, Kite tells you which ones to create.

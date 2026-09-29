---
sidebar_position: 4.5
---

# Secrets

Secrets keep API keys and other private values out of your flows. Put a key in a secret instead of pasting it into a block, and it isn't part of the flow when you export or share it.

Create secrets under _Secrets_ in your app. Use one in a [Send API Request](./blocks/actions/action_http_request.md) block's URL, headers, query parameters or body as `{{secrets.NAME}}`, for example a header `Authorization` with the value `Bearer {{secrets.OPENWEATHER_KEY}}`.

Secrets only work in API request blocks. Everywhere else, like in messages or logs, `{{secrets.NAME}}` fails, so a secret can't be sent or shown by accident. If a request fails, the error only shows the server the request went to, and secret values are replaced with `[secret]`. Names start with a letter or underscore and can contain letters, numbers and underscores.

Once saved, a value can't be read back, not even by you. To change it, enter a new one. Everyone who can edit your app can use its secrets in their blocks, so this protects keys from people you share flows with, not from your collaborators.

When you import a flow that uses secrets your app doesn't have, Kite tells you which ones to create.
